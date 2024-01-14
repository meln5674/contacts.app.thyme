package main_test

import (
	"os"
	"path/filepath"

	"github.com/onsi/biloba"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	app "github.com/meln5674/hypermedia-systems-templ"
	"github.com/meln5674/hypermedia-systems-templ/model"
)

var _ = Describe("App", func() {
	var b *biloba.Biloba
	BeforeEach(func() {
		b = biloba.ConnectToChrome(GinkgoT())
		tmpDir, err := os.MkdirTemp("", "")
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() {
			os.RemoveAll(tmpDir)
		})
		app := app.App{Contacts: model.At(filepath.Join(tmpDir, "db.json"))}
		stop := make(chan struct{})
		DeferCleanup(func() {
			close(stop)
		})
		go app.LoadAndRun("localhost:8080", stop)
	})
	Describe("Root path", func() {
		It("should redirect to /contacts", func() {
			b.Navigate("http://localhost:8080")
			Eventually(b.Location).Should(Equal("http://localhost:8080/contacts"))
		})
	})
})
