package main_test

import (
	"encoding/json"
	"os/exec"

	"github.com/concourse/semver-resource/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Check", func() {
	var checkCmd *exec.Cmd

	BeforeEach(func() {
		checkCmd = exec.Command(checkPath)
	})

	Context("when executed", func() {
		var request models.CheckRequest
		var response models.CheckResponse

		BeforeEach(func() {
			request = models.CheckRequest{}
			response = models.CheckResponse{}
		})

		JustBeforeEach(func() {
			stdin, err := checkCmd.StdinPipe()
			Ω(err).ShouldNot(HaveOccurred())

			session, err := gexec.Start(checkCmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("outputs an empty list", func() {
			Ω(response).Should(HaveLen(0))
		})
	})
})
