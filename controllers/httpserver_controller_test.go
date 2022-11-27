package controllers

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTP server controller", func() {
	const timeout = time.Second * 3
	const interval = time.Millisecond * 250

	It("Should start a server", func() {
		Eventually(func(g Gomega) {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/", jobServerAddr), nil)
			g.Expect(err).ShouldNot(HaveOccurred())
			resp, err := http.DefaultClient.Do(req.WithContext(ctx))
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(resp.StatusCode).Should(Equal(404))
		}, timeout, interval).Should(Succeed())
	})
})
