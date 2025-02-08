package cache_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"assignment_service/pkg/infra/cache"
)

func TestOrderedMap(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ordered map test suite")
}

var _ = Describe("Ordered map", func() {
	var (
		om cache.OrderedMap
	)

	BeforeEach(func() {
		om = cache.NewOrderedMap[string]()
	})

	Describe("Positive cases", func() {
		It("creates a new OrderedMap", func() {
			Expect(om).NotTo(BeNil())
		})

		It("adds and retrieves a key-value pair", func() {
			om.Set("testKey-1", "testVal-1")
			value, ok := om.Get("testKey-1")
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("testVal-1"))
		})

		It("updates an existing key", func() {
			om.Set("testKey-1", "testVal-1")
			om.Set("testKey-2", "testVal-2")
			om.Set("testKey-1", "testVal-3")

			val, ok := om.Get("testKey-1")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("testVal-3"))

			items := om.Items()
			Expect(items).To(HaveLen(2))
			Expect(items[0].Key).To(Equal("testKey-1"))
			Expect(items[1].Key).To(Equal("testKey-2"))
		})

		It("maintains insertion order", func() {
			om.Set("testKey-1", "testVal-1")
			om.Set("testKey-2", "testVal-2")
			om.Set("tel aviv", "testVal-3")

			items := om.Items()
			Expect(items).To(HaveLen(3))
			Expect(items[0].Key).To(Equal("testKey-1"))
			Expect(items[1].Key).To(Equal("testKey-2"))
			Expect(items[2].Key).To(Equal("tel aviv"))
		})

		It("deletes the key", func() {
			om.Set("testKey-1", "testVal-1")
			om.Set("testKey-2", "testVal-2")
			om.Set("tel aviv", "testVal-3")

			deleted := om.Delete("testKey-2")
			Expect(deleted).To(BeTrue())

			_, ok := om.Get("testKey-2")
			Expect(ok).To(BeFalse())

			items := om.Items()
			Expect(items).To(HaveLen(2))
			Expect(items[0].Key).To(Equal("testKey-1"))
			Expect(items[1].Key).To(Equal("tel aviv"))
		})

		It("retrieves all items", func() {
			om.Set("x", "1")
			om.Set("y", "2")
			om.Set("z", "3")

			items := om.Items()
			Expect(items).To(HaveLen(3))
			Expect(items[0].Key).To(Equal("x"))
			Expect(items[0].Value).To(Equal("1"))
			Expect(items[1].Key).To(Equal("y"))
			Expect(items[1].Value).To(Equal("2"))
			Expect(items[2].Key).To(Equal("z"))
			Expect(items[2].Value).To(Equal("3"))
		})

		It("inserts the same key multiple times ", func() {
			om.Set("repeat", "100")
			om.Set("repeat", "200")
			val, ok := om.Get("repeat")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("200"))
		})
	})

	Describe("Negative Cases", func() {
		It("returns 'not found' for a non-existent key", func() {
			_, ok := om.Get("non-exist")
			Expect(ok).To(BeFalse())
		})

		It("return false when deleting a non-existent key", func() {
			om.Set("testKey-1", "testVal-1")
			deleted := om.Delete("non-exist")
			Expect(deleted).To(BeFalse())

			value, ok := om.Get("testKey-1")
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("testVal-1"))
		})

		It("work correctly when empty", func() {
			items := om.Items()
			Expect(items).To(BeEmpty())

			_, ok := om.Get("non-exist")
			Expect(ok).To(BeFalse())
		})
	})
})
