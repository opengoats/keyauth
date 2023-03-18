package impl

import (
	"context"

	"github.com/opengoats/goat/exception"
	"github.com/opengoats/keyauth/apps/book"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newQueryBookRequest(r *book.QueryBookRequest) *queryBookRequest {
	return &queryBookRequest{
		r,
	}
}

type queryBookRequest struct {
	*book.QueryBookRequest
}

// 过滤条件
// 由于Mongodb支持嵌套，JSON，如何过滤嵌套嵌套里面的条件，使用.访问嵌套对象属性
func (r *queryBookRequest) FindFilter() bson.M {

	filter := bson.M{"status": bson.M{"$gt": 0}}

	if r.Keywords != "" {
		filter["$or"] = bson.A{
			bson.M{"data.book_name": bson.M{"$regex": r.Keywords, "$options": "im"}},
			bson.M{"data.author": bson.M{"$regex": r.Keywords, "$options": "im"}},
		}
	}
	return filter
}

// Find参数
func (r *queryBookRequest) FindOptions() *options.FindOptions {
	pageSize := int64(r.Page.PageSize)
	skip := int64(r.Page.PageSize) * int64(r.Page.PageNumber-1)

	opt := &options.FindOptions{
		// 排序：Order By create_at Desc
		Sort: bson.D{
			{Key: "create_at", Value: -1},
		},
		// 分页：limit 0,10 skip:0, limit:10
		Limit: &pageSize,
		Skip:  &skip,
	}

	return opt
}

// Save Object
func (s *service) save(ctx context.Context, ins *book.Book) error {
	if _, err := s.col.InsertOne(ctx, ins); err != nil {
		return err
	}
	return nil
}

// LIST, Query, 会很多条件(分页, 关键字, 条件过滤, 排序)
// 需要单独为其 做过滤参数构建
func (s *service) query(ctx context.Context, req *queryBookRequest) (*book.BookSet, error) {
	// SQL Where
	// FindFilter
	resp, err := s.col.Find(ctx, req.FindFilter(), req.FindOptions())
	if err != nil {
		return nil, exception.NewInternalServerError("find book error, error is %s", err)
	}

	set := book.NewBookSet()

	for resp.Next(ctx) {
		ins := book.NewDefaultBook()
		if err := resp.Decode(ins); err != nil {
			return nil, exception.NewInternalServerError("decode book error, error is %s", err)
		}
		set.Add(ins)
	}

	set.Total = int64(len(set.Items))
	return set, nil
}

// GET, Describe
// filter 过滤器(Collection),类似于MYSQL Where条件
// 调用Decode方法来进行 反序列化  bytes ---> Object (通过BSON Tag)
func (s *service) get(ctx context.Context, id string) (*book.Book, error) {
	filter := bson.M{"_id": id}
	filter["$and"] = bson.A{
		bson.M{"status": bson.M{"$gt": 0}},
	}

	ins := book.NewDefaultBook()
	if err := s.col.FindOne(ctx, filter).Decode(ins); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, exception.NewNotFound("book %s not found", id)
		}

		return nil, exception.NewInternalServerError("find book %s error, %s", id, err)
	}

	return ins, nil
}

// UpdateByID, 通过主键来更新对象
func (s *service) update(ctx context.Context, ins *book.Book) error {
	// SQL update obj(SET f=v,f=v) where id=?
	// s.col.UpdateOne(ctx, filter(), ins)
	data := bson.M{"$set": ins}
	if _, err := s.col.UpdateByID(ctx, ins.Id, data); err != nil {
		return exception.NewInternalServerError("inserted book(%s) document error, %s", ins.Data.BookName, err)
	}

	return nil
}
