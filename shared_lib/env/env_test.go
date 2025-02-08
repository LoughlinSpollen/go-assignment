package config_test

import (
	config "assignment/lib/shared_lib/env"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInfraEnv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "env config unit test suite")
}

var _ = Describe("environment variables", func() {
	Describe("Reading int64 env variables", func() {
		int64Val := rand.Int64()
		Context("when the environment variable is a valid int64 number", func() {
			BeforeEach(func() {
				int64Str := strconv.FormatInt(int64Val, 10)
				os.Setenv("ENV_VAT_INT64", int64Str)
			})
			AfterEach(func() {
				os.Unsetenv("ENV_VAT_INT64")
			})
			It("return the env var as a int64", func() {
				result := config.WithDefaultInt64("ENV_VAT_INT64", "0")
				Expect(result).Should(Equal(int64Val))
			})
		})

		Context("when the environment variable is an invalid int64 number", func() {
			BeforeEach(func() {
				os.Setenv("ENV_VAT_INT64", "*")
			})
			AfterEach(func() {
				os.Unsetenv("ENV_VAT_INT64")
			})

			It("panics", func() {
				defer func() {
					if r := recover(); r != nil {
						_, isError := r.(error)
						Expect(isError).To(BeTrue())
					} else {
						Fail("expected panic")
					}
				}()

				result := config.WithDefaultInt64("ENV_VAT_INT64", "0")
				Expect(result).Should(BeEmpty())
			})
		})

		Context("when the environment variable is empty", func() {
			BeforeEach(func() {
				os.Unsetenv("ENV_VAT_INT64")
			})

			It("returns the default value as an int64", func() {
				int64Default := rand.Int64()
				int64DefaultStr := strconv.FormatInt(int64Default, 10)
				result := config.WithDefaultInt64("ENV_VAT_INT64", int64DefaultStr)
				Expect(result).Should(Equal(int64Default))
			})
		})
	})

	Describe("Reading int env variables", func() {
		intVal := rand.Int()
		Context("when the environment variable is a valid int number", func() {
			BeforeEach(func() {
				os.Setenv("ENV_VAT_INT", strconv.Itoa(intVal))
			})
			AfterEach(func() {
				os.Unsetenv("ENV_VAT_INT")
			})
			It("returns the env variable value as an int", func() {
				result := config.WithDefaultInt("ENV_VAT_INT", "0")
				Expect(result).Should(Equal(intVal))
			})
		})

		Context("when the environment variable is not a number", func() {
			BeforeEach(func() {
				os.Setenv("ENV_VAT_INT", "*")
			})
			AfterEach(func() {
				os.Unsetenv("ENV_VAT_INT")
			})

			It("should panic", func() {
				defer func() {
					if r := recover(); r != nil {
						_, isError := r.(error)
						Expect(isError).To(BeTrue())
					} else {
						Fail("expected panic")
					}
				}()

				result := config.WithDefaultInt("ENV_VAT_INT", "8080")
				Expect(result).Should(BeEmpty())
			})
		})

		Context("when the environment variable is empty", func() {
			BeforeEach(func() {
				os.Unsetenv("ENV_VAT_INT")
			})

			It("returns the default value", func() {
				result := config.WithDefaultInt("ENV_VAT_INT", "1")
				Expect(result).Should(Equal(1))
			})
		})
	})

	Describe("Reading string env variables", func() {
		Context("when the string variable is a valid environment", func() {
			BeforeEach(func() {
				os.Setenv("ENV_VAT_STR", "expected-string")
			})
			AfterEach(func() {
				os.Unsetenv("ENV_VAT_STR")
			})
			It("return the env variable value as a string", func() {
				result := config.WithDefaultString("ENV_VAT_STR", "default-string")
				Expect(result).Should(Equal("expected-string"))
			})
		})

		Context("when the string environment variable is not set", func() {
			It("return the default variable value as a string", func() {
				result := config.WithDefaultString("ENV_VAT_STR", "default-string")
				Expect(result).Should(Equal("default-string"))
			})
		})
	})
})
