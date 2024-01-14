package main_test

import (
	"testing"

	"github.com/onsi/biloba"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHypermediaSystemsTempl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HypermediaSystemsTempl Suite")
}

var _ = BeforeSuite(func() {
	biloba.SpinUpChrome(GinkgoT())
})
