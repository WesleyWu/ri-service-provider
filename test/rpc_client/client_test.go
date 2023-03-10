package rpc_client

import (
	"context"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"fmt"
	"github.com/WesleyWu/gowing/protobuf/gwtypes"
	"github.com/WesleyWu/gowing/rpc/dubbogo"
	"github.com/WesleyWu/gowing/util/gwwrapper"
	"github.com/WesleyWu/ri-service-provider/proto/video_collection"
	"github.com/gogf/gf/contrib/trace/jaeger/v2"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	ctx                   = gctx.New()
	videoCollectionClient = new(proto_video_collection.VideoCollectionClientImpl)
	ServiceName           = "VideoCollectionTest"
	JaegerUdpEndpoint     = "localhost:6831"
)

func init() {
	err := startDubboConsumer(ctx)
	if err != nil {
		panic(err)
	}
}

func TestCount(t *testing.T) {
	req := &proto_video_collection.VideoCollectionCountReq{
		Id:          nil,
		Name:        gwwrapper.AnyString("推荐视频集"),
		ContentType: gwwrapper.AnyUInt32Slice([]uint32{1, 2}),
		FilterType:  nil,
		Count:       nil,
		IsOnline:    nil,
		CreatedAt:   nil,
		UpdatedAt:   nil,
	}
	fmt.Println(gjson.MustEncodeString(req))
	res, err := videoCollectionClient.Count(ctx, req)
	if err != nil {
		panic(err)
	}
	fmt.Println(gjson.MustEncodeString(res))
	assert.Equal(t, int32(2), res.Total)
}

func TestList(t *testing.T) {
	tp, err := jaeger.Init(ServiceName, JaegerUdpEndpoint)
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	defer tp.Shutdown(ctx)
	ctx, span := gtrace.NewSpan(ctx, "ClientList")
	defer span.End()
	req := &proto_video_collection.VideoCollectionListReq{
		Id:          nil,
		Name:        gwwrapper.AnyCondition(gwtypes.OperatorType_Like, gwtypes.MultiType_Exact, gwtypes.WildcardType_Contains, gwwrapper.AnyString("每日")),
		ContentType: gwwrapper.AnyUInt32Slice([]uint32{1, 2}),
		FilterType:  nil,
		Count:       gwwrapper.AnyCondition(gwtypes.OperatorType_GT, gwtypes.MultiType_Exact, gwtypes.WildcardType_None, gwwrapper.AnyUInt32(1)),
		IsOnline:    nil,
		CreatedAt:   nil,
		UpdatedAt:   nil,
	}
	fmt.Println(gjson.MustEncodeString(req))
	res, err := videoCollectionClient.List(ctx, req)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, int32(1), *res.Total)
	fmt.Println(gjson.MustEncodeString(res))
	res, err = videoCollectionClient.List(ctx, req)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, int32(1), *res.Total)
}

func TestListBenchmark(t *testing.T) {
	tp, err := jaeger.Init(ServiceName, JaegerUdpEndpoint)
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	defer tp.Shutdown(ctx)
	req := &proto_video_collection.VideoCollectionListReq{
		Id:          nil,
		Name:        gwwrapper.AnyCondition(gwtypes.OperatorType_Like, gwtypes.MultiType_Exact, gwtypes.WildcardType_Contains, gwwrapper.AnyString("每日")),
		ContentType: gwwrapper.AnyUInt32Slice([]uint32{1, 2}),
		FilterType:  nil,
		Count:       gwwrapper.AnyCondition(gwtypes.OperatorType_GT, gwtypes.MultiType_Exact, gwtypes.WildcardType_None, gwwrapper.AnyUInt32(1)),
		IsOnline:    nil,
		CreatedAt:   nil,
		UpdatedAt:   nil,
	}
	fmt.Println(gjson.MustEncodeString(req))
	timeStart := gtime.Now()
	benchCount := 10000
	for i := 0; i < benchCount; i++ {
		ctx, span := gtrace.NewSpan(ctx, "ClientBenchmark"+gconv.String(i))
		res, err := videoCollectionClient.List(ctx, req)
		if err != nil {
			panic(err)
		}
		assert.Equal(t, int32(1), *res.Total)
		if (i+1)%1000 == 0 {
			g.Log().Infof(ctx, "called %d times", i+1)
		}
		span.End()
	}
	timeEnd := gtime.Now()
	millisElapsed := timeEnd.UnixMilli() - timeStart.UnixMilli()
	cps := float64(benchCount) * 1000 / float64(millisElapsed)
	g.Log().Infof(ctx, "RPC Calls per seconds for VideoCollection.List: %.2f", cps)
}

func TestCreate(t *testing.T) {
	res, err := videoCollectionClient.Create(ctx, &proto_video_collection.VideoCollectionCreateReq{
		Id:          gwwrapper.WrapString("87104859-5598"),
		Name:        gwwrapper.WrapString("特别长的名称特别长的名称特别长的名称特别长的"),
		ContentType: gwwrapper.WrapInt32(3),
		FilterType:  gwwrapper.WrapInt32(4),
		Count:       gwwrapper.WrapUInt32(401),
		IsOnline:    gwwrapper.WrapBool(true),
	})
	if err != nil {
		panic(err)
	}
	g.Log().Infof(ctx, gjson.MustEncodeString(res))
}

func startDubboConsumer(ctx context.Context) error {
	registryId := g.Cfg().MustGet(ctx, "rpc.registry.id", "nacosRegistry").String()
	registryProtocol := g.Cfg().MustGet(ctx, "rpc.registry.protocol", "nacos").String()
	registryAddress := g.Cfg().MustGet(ctx, "rpc.registry.address", "127.0.0.1:8848").String()
	registryNamespace := g.Cfg().MustGet(ctx, "rpc.registry.namespace", "public").String()
	development := g.Cfg().MustGet(ctx, "server.debug", "true").Bool()
	loggerStdout := g.Cfg().MustGet(ctx, "logger.stdout", "true").Bool()
	loggerPath := g.Cfg().MustGet(ctx, "rpc.consumer.logDir").String()
	if g.IsEmpty(loggerPath) {
		loggerPath = g.Cfg().MustGet(ctx, "logger.path", "./data/log/gf-app").String()
	}
	loggerFileName := g.Cfg().MustGet(ctx, "rpc.consumer.logFile", "consumer.log").String()
	loggerLevel := g.Cfg().MustGet(ctx, "rpc.consumer.logLevel", "warn").String()

	dubbogo.AddConsumerReference(
		&dubbogo.ConsumerReference{
			ClientImplStructName: "VideoCollectionClientImpl",
			Service:              videoCollectionClient,
			Protocol:             "tri",
		})
	return dubbogo.StartConsumers(ctx,
		&dubbogo.Registry{
			Id:        registryId,
			Type:      registryProtocol,
			Address:   registryAddress,
			Namespace: registryNamespace,
		},
		&dubbogo.ConsumerOption{
			CheckProviderExists: true,
			TimeoutSeconds:      180,
		},
		&dubbogo.LoggerOption{
			Development: development,
			Stdout:      loggerStdout,
			LogDir:      loggerPath,
			LogFileName: loggerFileName,
			Level:       loggerLevel,
		})
}
