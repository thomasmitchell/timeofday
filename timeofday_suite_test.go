package timeofday_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTimeofday(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Timeofday Suite")
}
