package service

import (
	dto2 "api/app/books/dto"
	"api/app/books/entitie"
	"api/pkg/common"
	"api/pkg/dto"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"math"
)

type ArticleService struct {
	article *mongo.Collection
	log     *zap.Logger
}

func NewArticleService(app *common.App) *ArticleService {
	return &ArticleService{
		article: app.Db.Collection("article"),
		log:     app.Log,
	}
}

func (artSrv ArticleService) FindById(id string) (*entitie.Article, error) {
	articleEntity := &entitie.Article{}
	result := artSrv.article.FindOne(context.TODO(), bson.M{"_id": id})
	err := result.Err()
	if err != nil {
		artSrv.log.Warn("Cannot findById(id)", zap.String("id", id), zap.Error(err))
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, common.NotFound
		}
		return nil, err
	}
	err = result.Decode(articleEntity)
	if err != nil {
		artSrv.log.Warn("error occurs when Decode(articleEntity)", zap.String("id", id), zap.Error(err))
	}
	return articleEntity, err
}

func (artSrv *ArticleService) List(catalogId string, pageRequest *dto2.PageRequest) (*dto2.ArticlePageResponse, error) {
	var results = new([]*entitie.Article) //分配空间返回地址
	findOpt := options.Find()
	findOpt.SetLimit(int64(pageRequest.PageSize))
	findOpt.SetSkip(int64((pageRequest.Page - 1) * pageRequest.PageSize))
	findOpt.SetProjection(bson.M{"content": 0}) //不包含content内容

	cursor, err := artSrv.article.Find(context.TODO(), bson.M{"catalogId": catalogId}, findOpt)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	if err != nil {
		artSrv.log.Warn("An error occurs while getting a list of articles", zap.Error(err))
		return nil, err
	}

	err = cursor.All(context.TODO(), results)
	if err != nil {
		artSrv.log.Warn("An error occurs while decoding a book article", zap.Error(err))
		return nil, err
	}
	//第二种方法，迭代每一个对象处理
	//for cursor.Next(context.TODO()) {
	//	var article *entitie.Article
	//	err := cursor.Decode(&article)
	//	if err != nil {
	//		artSrv.log.Warn("An error occurs while decoding a book article", zap.Error(err))
	//		return nil, err
	//	}
	//	results = append(results, article)
	//}

	//查询总条数
	countOptions := &options.CountOptions{}
	count, err := artSrv.article.CountDocuments(context.TODO(), bson.M{"catalogId": catalogId}, countOptions)
	if err != nil {
		artSrv.log.Warn("An error occurs while counting articles", zap.Error(err))
		return nil, err
	}

	resp := &dto2.ArticlePageResponse{
		Page:         pageRequest.Page,
		TotalPage:    int32(math.Ceil(float64(count) / float64(pageRequest.PageSize))),
		PageSize:     pageRequest.PageSize,
		TotalRecords: int32(count),
		Result: dto.Result{
			Payload: results,
		},
	}

	return resp, nil
}

func (artSrv *ArticleService) Search(pageRequest *dto2.PageRequest) (*dto2.ArticlePageResponse, error) {
	//options.
	findOpt := options.Find()
	findOpt.SetLimit(int64(pageRequest.PageSize))
	findOpt.SetSkip(int64((pageRequest.Page - 1) * pageRequest.PageSize))
	findOpt.SetProjection(bson.M{"content": 0})

	count, err := artSrv.article.CountDocuments(context.TODO(), bson.M{"$text": bson.M{"$search": pageRequest.Search}})
	if err != nil {
		artSrv.log.Warn("An error occurs while getting the number of articles", zap.Error(err))
		return nil, err
	}

	cursor, err := artSrv.article.Find(context.TODO(), bson.M{"$text": bson.M{"$search": pageRequest.Search}}, findOpt)
	if err != nil {
		artSrv.log.Warn("An error occurs while counting articles", zap.Error(err))
		return nil, err
	}
	var results = new([]*entitie.Article) //分配空间返回地址
	err = cursor.All(context.TODO(), results)
	if err != nil {
		artSrv.log.Warn("An error occurs while converting articles", zap.Error(err))
		return nil, err
	}

	resp := &dto2.ArticlePageResponse{
		Page:         pageRequest.Page,
		TotalPage:    int32(math.Ceil(float64(count) / float64(pageRequest.PageSize))),
		PageSize:     pageRequest.PageSize,
		TotalRecords: int32(count),
		Result: dto.Result{
			Payload: results,
		},
	}

	return resp, nil
}
