package tlsconfig_test

import (
	"crypto/tls"
	"crypto/x509"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/paraphernalia/test/certtest"

	"github.com/pivotal-cf/paraphernalia/secure/tlsconfig"
)

var _ = Describe("generating TLS configurations", func() {
	var (
		config  *tls.Config
		tlsOpts []tlsconfig.TLSOption
	)

	ItCanUseInternalPivotalDefaults := func() {
		Describe("with internal Pivotal defaults", func() {
			BeforeEach(func() {
				tlsOpts = []tlsconfig.TLSOption{
					tlsconfig.WithPivotalDefaults(),
				}
			})

			It("makes sure that the server is the source of truth for cipher suites", func() {
				Expect(config.PreferServerCipherSuites).To(BeTrue())
			})

			It("enforces the use of TLS 1.2", func() {
				Expect(config.MinVersion).To(Equal(uint16(tls.VersionTLS12)))
			})

			It("uses approved cipher suites", func() {
				Expect(config.CipherSuites).To(Equal([]uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				}))
			})

			It("uses approved curves", func() {
				Expect(config.CurvePreferences).To(ConsistOf(tls.CurveP384))
			})
		})
	}

	ItCanBeAssignedAnIdentity := func() {
		Describe("with a client identity", func() {
			var tlsCert tls.Certificate

			BeforeEach(func() {
				ca, err := certtest.BuildCA("tlsconfig")
				Expect(err).NotTo(HaveOccurred())

				cert, err := ca.BuildSignedCertificate("tlsconfig")
				Expect(err).NotTo(HaveOccurred())

				tlsCert, err = cert.TLSCertificate()
				Expect(err).NotTo(HaveOccurred())

				tlsOpts = []tlsconfig.TLSOption{
					tlsconfig.WithIdentity(tlsCert),
				}
			})

			It("sets the certificates", func() {
				Expect(config.Certificates).To(ConsistOf(tlsCert))
			})
		})
	}

	Describe("server configurations", func() {
		var (
			serverOpts []tlsconfig.ServerOption
		)

		JustBeforeEach(func() {
			config = tlsconfig.Build(tlsOpts...).Server(serverOpts...)
		})

		ItCanUseInternalPivotalDefaults()
		ItCanBeAssignedAnIdentity()

		Describe("with client authentication", func() {
			var pool *x509.CertPool

			BeforeEach(func() {
				ca, err := certtest.BuildCA("tlsconfig")
				Expect(err).NotTo(HaveOccurred())

				pool, err = ca.CertPool()
				Expect(err).NotTo(HaveOccurred())

				serverOpts = []tlsconfig.ServerOption{
					tlsconfig.WithClientAuthentication(pool),
				}
			})

			It("makes sure we require client authentication", func() {
				Expect(config.ClientAuth).To(Equal(tls.RequireAndVerifyClientCert))
			})

			It("sets the client authority", func() {
				Expect(config.ClientCAs).NotTo(BeNil())
			})
		})
	})

	Describe("client configurations", func() {
		var (
			clientOpts []tlsconfig.ClientOption
		)

		JustBeforeEach(func() {
			config = tlsconfig.Build(tlsOpts...).Client(clientOpts...)
		})

		ItCanUseInternalPivotalDefaults()
		ItCanBeAssignedAnIdentity()

		Describe("with authority", func() {
			var pool *x509.CertPool

			BeforeEach(func() {
				ca, err := certtest.BuildCA("tlsconfig")
				Expect(err).NotTo(HaveOccurred())

				pool, err = ca.CertPool()
				Expect(err).NotTo(HaveOccurred())

				clientOpts = []tlsconfig.ClientOption{
					tlsconfig.WithAuthority(pool),
				}
			})

			It("sets the client authority", func() {
				Expect(config.RootCAs).NotTo(BeNil())
			})
		})
	})

	Describe("configuration modification", func() {
		It("does not affect other configurations", func() {
			base := tlsconfig.Build()
			client := base.Client()

			ca, err := certtest.BuildCA("tlsconfig")
			Expect(err).NotTo(HaveOccurred())

			pool, err := ca.CertPool()
			Expect(err).NotTo(HaveOccurred())

			server := base.Server(
				tlsconfig.WithClientAuthentication(pool),
			)

			Expect(client.ClientAuth).NotTo(Equal(server.ClientAuth))
		})
	})
})
