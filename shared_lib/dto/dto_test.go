package dto_test

import (
	"assignment/lib/shared_lib/domain"
	"context"
	"encoding/json"
	"fmt"
	dto "shared_lib/dto"
	"testing"

	validator "github.com/go-playground/validator/v10"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDTOAdapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DTO adaper test suite")
}

var _ = Describe("DTOAdapter", func() {
	var (
		dtoAdapter dto.DtoAdapter
		ctx        context.Context
	)
	BeforeEach(func() {
		dtoAdapter = dto.DtoAdapter{
			Validator: validator.New(),
		}
		err := dtoAdapter.RegisterCommandTypeValidation()
		Expect(err).NotTo(HaveOccurred())
		ctx = context.Background()
	})

	Describe("ToDTO", func() {
		It("converts a Command to a DTO successfully", func() {
			cmd := domain.Command{
				Name: domain.AddItem,
				Data: []domain.DataEntry{{Key: "testKey", Value: "testValue"}},
			}

			result, err := dtoAdapter.ToDTO(ctx, &cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			var decoded dto.CommandDTO
			err = json.Unmarshal(result, &decoded)
			Expect(err).NotTo(HaveOccurred())
			Expect(decoded.Name).To(Equal(cmd.Name))
			Expect(decoded.Data).To(HaveLen(1))
			Expect(decoded.Data[0].Key).To(Equal(cmd.Data[0].Key))
			Expect(decoded.Data[0].Value).To(Equal(cmd.Data[0].Value))
		})
	})

	Describe("FromDTO", func() {
		It("decodes a valid DTO successfully", func() {
			cmd := dto.CommandDTO{
				Name: domain.DeleteItem,
				Data: []dto.DataEntryDTO{{Key: "testKey", Value: "testValue"}},
			}
			jsonBytes, _ := json.Marshal(cmd)

			result, err := dtoAdapter.FromDTO(ctx, jsonBytes)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal(cmd.Name))
			Expect(result.Data).To(HaveLen(1))
			Expect(result.Data[0].Key).To(Equal(cmd.Data[0].Key))
			Expect(result.Data[0].Value).To(Equal(cmd.Data[0].Value))
		})

		Context("negative cases", func() {
			It("returns an error for missing data", func() {
				cmd := dto.CommandDTO{Name: domain.AddItem}
				jsonBytes, _ := json.Marshal(cmd)

				_, err := dtoAdapter.FromDTO(ctx, jsonBytes)
				Expect(err).To(HaveOccurred())
			})

			It("returns an error for invalid JSON", func() {
				invalidJSON := []byte("{invalid_json}")

				_, err := dtoAdapter.FromDTO(ctx, invalidJSON)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ToErrorDTO", func() {
		It("converts an error to a DTO successfully", func() {
			errTest := fmt.Errorf("test error")
			result, err := dtoAdapter.ToErrorDTO(ctx, errTest)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			var decoded dto.ErrorDTO
			err = json.Unmarshal(result, &decoded)
			Expect(err).NotTo(HaveOccurred())
			Expect(decoded.Message).To(Equal(errTest.Error()))
		})
	})

	Describe("ValidateCommandType", func() {
		Context("AddItem command", func() {
			It("should pass validation if the Data entry has a Key and Value", func() {
				cmd := dto.CommandDTO{
					Name: domain.AddItem,
					Data: []dto.DataEntryDTO{
						{Key: "testKey", Value: "testValue"},
					},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should fail validation if the Data entry is missing the Key", func() {
				cmd := dto.CommandDTO{
					Name: domain.AddItem,
					Data: []dto.DataEntryDTO{
						{Key: "", Value: "testValue"},
					},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).To(HaveOccurred())
			})

			It("should fail validation if the Data entry is missing the Value", func() {
				cmd := dto.CommandDTO{
					Name: domain.AddItem,
					Data: []dto.DataEntryDTO{
						{Key: "testKey", Value: ""},
					},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).To(HaveOccurred())
			})

			It("should fail validation if the Data entry is missing Key and Value", func() {
				cmd := dto.CommandDTO{
					Name: domain.AddItem,
					Data: []dto.DataEntryDTO{
						{Key: "", Value: ""},
					},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).To(HaveOccurred())
			})

			It("should fail validation if no Data entry is present", func() {
				cmd := dto.CommandDTO{
					Name: domain.AddItem,
					Data: []dto.DataEntryDTO{},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("GetItem command", func() {
			It("should pass validation if the Data entry has a Key", func() {
				cmd := dto.CommandDTO{
					Name: domain.GetItem,
					Data: []dto.DataEntryDTO{
						{Key: "testKey"},
					},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should fail validation if the Data entry is missing the Key", func() {
				cmd := dto.CommandDTO{
					Name: domain.GetItem,
					Data: []dto.DataEntryDTO{
						{Key: "", Value: "testValue"},
					},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).To(HaveOccurred())
			})
			It("should fail validation if no Data entry is present", func() {
				cmd := dto.CommandDTO{
					Name: domain.GetItem,
					Data: []dto.DataEntryDTO{},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when Name is DeleteItem", func() {
			It("should pass validation if the Data entry has the Key", func() {
				cmd := dto.CommandDTO{
					Name: domain.DeleteItem,
					Data: []dto.DataEntryDTO{
						{Key: "testKey"},
					},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should fail validation if the Data entry is missing the Key", func() {
				cmd := dto.CommandDTO{
					Name: domain.DeleteItem,
					Data: []dto.DataEntryDTO{
						{Key: ""},
					},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).To(HaveOccurred())
			})

			It("should fail validation if no Data entry is present", func() {
				cmd := dto.CommandDTO{
					Name: domain.DeleteItem,
					Data: []dto.DataEntryDTO{},
				}
				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when Name is GetAllItems", func() {
			It("should pass validation regardless of Data content", func() {
				cmd := dto.CommandDTO{
					Name: domain.GetAllItems,
					Data: []dto.DataEntryDTO{
						{Key: "", Value: ""},
					},
				}

				err := dtoAdapter.Validator.Struct(cmd)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
