package main_test

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/onsi/biloba"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	app "github.com/meln5674/contacts.app.thyme"
	"github.com/meln5674/contacts.app.thyme/model"
)

var _ = Describe("App", func() {
	var b *biloba.Biloba
	app := app.App{}
	BeforeEach(func() {
		b = biloba.ConnectToChrome(GinkgoT())
		tmpDir, err := os.MkdirTemp("", "")
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() {
			os.RemoveAll(tmpDir)
		})
		app.Contacts = model.At(filepath.Join(tmpDir, "db.json"))
		stop := make(chan struct{})
		DeferCleanup(func() {
			close(stop)
		})
		go func() {
			defer GinkgoRecover()
			err := app.LoadAndRun("localhost:8080", stop)
			Expect(stop).To(BeClosed())
			Expect(err).To(MatchError(http.ErrServerClosed))
		}()
	})
	Describe("Root path", func() {
		BeforeEach(func() {
			b.Navigate("http://localhost:8080")
		})
		It("should redirect to /contacts", func() {
			Eventually(b.Location).Should(Equal("http://localhost:8080/contacts"))
		})
	})
	Describe("Contacts page", func() {
		BeforeEach(func() {
			b.Navigate("http://localhost:8080/contacts")
		})
		It("should show the title", func() {
			Eventually("body h1").Should(b.HaveInnerText("contacts.app\nA Demo Contacts Application"))
		})
		It("should have an empty table", func() {
			Eventually("body table").Should(b.Exist())
			Expect("body table thead tr").To(b.HaveCount(1))
			Expect("body table tbody tr").To(b.HaveCount(0))
		})
		Describe("Add contact button", func() {
			It("Should go to the edit page", func() {
				Eventually(`a[href="/contacts/new"`).Should(b.Exist())
				b.Click(`a[href="/contacts/new"`)
				Eventually(b.Location).Should(Equal("http://localhost:8080/contacts/new"))
			})
		})
		When("There is a contact", func() {
			BeforeEach(func() {
				saved, err := app.Contacts.Save(&model.Contact{
					First: "first",
					Last:  "last",
					Email: "email@example.com",
					Phone: "555-123-4567",
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(saved).To(BeTrue(), "Contact was not saved")
				// Need to renavigate since the update isn't automatic
				b.Navigate("http://localhost:8080/contacts")
			})
			It("Should show it", func() {
				Eventually("body table tbody tr").Should(b.HaveCount(1))
				Expect("body table tbody tr td:nth-child(1)").Should(b.HaveInnerText("first"))
				Expect("body table tbody tr td:nth-child(2)").Should(b.HaveInnerText("last"))
				Expect("body table tbody tr td:nth-child(3)").Should(b.HaveInnerText("555-123-4567"))
				Expect("body table tbody tr td:nth-child(4)").Should(b.HaveInnerText("email@example.com"))
			})
			It("Should show 1 total contacts", func() {
				Eventually("body > main > p > span").Should(b.HaveInnerText("(1 total Contacts)"))
			})
		})
	})
})
