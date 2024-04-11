package e2e

import (
	"fmt"
	"time"

	"gitbub.com/wbuntu/free-ask-bot/e2e/cases"
	"gitbub.com/wbuntu/free-ask-bot/internal/model"
	"gitbub.com/wbuntu/free-ask-bot/internal/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster", func() {
	It("CheckServer", func(ctx SpecContext) {
		var err error
		for i := 0; i < 3; i++ {
			request := &model.BaseRequest{}
			response := &model.BaseResponse{}
			err = httpClient.GET(
				ctx,
				"/readyz",
				request,
				response,
			)
			if err == nil {
				break
			}
			GinkgoWriter.Printf("server error: %s\n", err)
			time.Sleep(time.Second * 10)
		}
		Expect(err).To(BeNil())
	})
	var e2eClusterID string
	It("CreateCluster", func(ctx SpecContext) {
		request := cases.StandardCreateClusterRequest
		response := &model.CreateClusterResp{}
		err := httpClient.POST(
			ctx,
			"/api/v1.0/clusters",
			request,
			response,
		)
		Expect(err).To(BeNil())
		Expect(response.Data).To(HavePrefix("cluster"))
		GinkgoWriter.Printf("cluster created: %s\n", response.Data)
		// 保存集群ID
		e2eClusterID = response.Data
	})
	It("ListCluster", func(ctx SpecContext) {
		request := &model.ListClusterReq{
			PageNo:   1,
			PageSize: 10,
		}
		response := &model.ListClusterResp{}
		err := httpClient.GET(
			ctx,
			"/api/v1.0/clusters",
			request,
			response,
		)
		Expect(err).To(BeNil())
		Expect(response.Data).NotTo(BeEmpty())
		Expect(response.TotalCount).NotTo(BeZero())
		GinkgoWriter.Printf("cluster totalCount: %d\n", response.TotalCount)
	})
	It("GetCluster", func(ctx SpecContext) {
		Eventually(func(g Gomega) {
			request := &model.GetClusterReq{}
			response := &model.GetClusterResp{}
			err := httpClient.GET(
				ctx,
				fmt.Sprintf("/api/v1.0/clusters/%s", e2eClusterID),
				request,
				response,
			)
			g.Expect(err).To(BeNil())
			g.Expect(response.Data.Status).To(Equal(storage.ClusterStatusRunning.String()))
			GinkgoWriter.Printf("cluster running: %s\n", e2eClusterID)
		}, time.Minute*120, time.Minute*3).Should(Succeed())
	})
	It("DeleteCluster", func(ctx SpecContext) {
		request := &model.BaseRequest{}
		response := &model.BaseResponse{}
		err := httpClient.DELETE(
			ctx,
			fmt.Sprintf("/api/v1.0/clusters/%s", e2eClusterID),
			request,
			response,
		)
		Expect(err).To(BeNil())
	})
})
