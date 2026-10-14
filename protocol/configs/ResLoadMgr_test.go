package configs

import (
	"mocktest/mocks"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestSingleLoadRes(t *testing.T) {
	var start bool = false

	ctrl := gomock.NewController(t)

	res1 := mocks.NewMockIResMgr(ctrl)
	res1.EXPECT().GetFileName().AnyTimes().Return("resource_1")
	res1.EXPECT().GetFilePath().AnyTimes().Return("/path/to/resource_1")
	res1.EXPECT().Load(start).AnyTimes().Return(nil)
	res1.EXPECT().PrintData().AnyTimes().Return("this is res1 print")

	res2 := mocks.NewMockIResMgr(ctrl)
	res2.EXPECT().GetFileName().AnyTimes().Return("resource_2")
	res2.EXPECT().GetFilePath().AnyTimes().Return("/path/to/resource_2")
	res2.EXPECT().Load(start).AnyTimes().Return(nil)
	res2.EXPECT().PrintData().AnyTimes().Return("this is res2 print")

	resLoader := GetResLoadMgr()

	resCallBack1 := mocks.NewMockIResCallbackMgr(ctrl)
	resCallBack1.EXPECT().GetName().AnyTimes().Return("callback_1")
	resCallBack1.EXPECT().Check(gomock.Any(), start).AnyTimes().Return(nil)
	resCallBack1.EXPECT().Load(start).AnyTimes()

	resCallBack2 := mocks.NewMockIResCallbackMgr(ctrl)
	resCallBack2.EXPECT().GetName().AnyTimes().Return("callback_2")
	resCallBack2.EXPECT().Check(gomock.Any(), start).AnyTimes().Return(nil)
	resCallBack2.EXPECT().Load(start).AnyTimes()

	resLoader.AddRes(res1.GetFileName(), res1)
	resLoader.AddRes(res2.GetFileName(), res2)
	resLoader.AddCallback(res2.GetFileName(), resCallBack2, resCallBack1)
	resLoader.AddCallback(res1.GetFileName(), resCallBack1, resCallBack2)
	t.Logf("-------------------case1--------------------------")
	resLoader.LoadAll(start)

	t.Logf("-------------------case2--------------------------")
	resLoader.Load(res2.GetFileName())

	t.Logf("-------------------case3--------------------------")
	resLoader.Load(res1.GetFileName())
}
