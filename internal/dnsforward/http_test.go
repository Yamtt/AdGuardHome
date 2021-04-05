package dnsforward

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsfilter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestData(t *testing.T, casesFileName string, cases interface{}) {
	t.Helper()

	var f *os.File
	f, err := os.Open(filepath.Join("testdata", casesFileName))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	err = json.NewDecoder(f).Decode(cases)
	require.NoError(t, err)
}

const jsonExt = ".json"

func TestDNSForwardHTTTP_handleGetConfig(t *testing.T) {
	filterConf := &dnsfilter.Config{
		SafeBrowsingEnabled:   true,
		SafeBrowsingCacheSize: 1000,
		SafeSearchEnabled:     true,
		SafeSearchCacheSize:   1000,
		ParentalCacheSize:     1000,
		CacheTime:             30,
	}
	forwardConf := ServerConfig{
		UDPListenAddrs: []*net.UDPAddr{},
		TCPListenAddrs: []*net.TCPAddr{},
		FilteringConfig: FilteringConfig{
			ProtectionEnabled: true,
			UpstreamDNS:       []string{"8.8.8.8:53", "8.8.4.4:53"},
		},
		ConfigModified: func() {},
	}
	s := createTestServer(t, filterConf, forwardConf)
	require.Nil(t, s.Start())
	t.Cleanup(func() {
		require.Nil(t, s.Stop())
	})

	defaultConf := s.conf

	w := httptest.NewRecorder()

	testCases := []struct {
		conf func() ServerConfig
		name string
	}{{
		conf: func() ServerConfig {
			return defaultConf
		},
		name: "all_right",
	}, {
		conf: func() ServerConfig {
			conf := defaultConf
			conf.FastestAddr = true

			return conf
		},
		name: "fastest_addr",
	}, {
		conf: func() ServerConfig {
			conf := defaultConf
			conf.AllServers = true

			return conf
		},
		name: "parallel",
	}}

	var data map[string]json.RawMessage
	loadTestData(t, t.Name()+jsonExt, &data)

	for _, tc := range testCases {
		caseWant, ok := data[tc.name]
		require.True(t, ok)

		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(w.Body.Reset)

			s.conf = tc.conf()
			s.handleGetConfig(w, nil)

			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, string(caseWant), w.Body.String())
		})
	}
}

func TestDNSForwardHTTTP_handleSetConfig(t *testing.T) {
	filterConf := &dnsfilter.Config{
		SafeBrowsingEnabled:   true,
		SafeBrowsingCacheSize: 1000,
		SafeSearchEnabled:     true,
		SafeSearchCacheSize:   1000,
		ParentalCacheSize:     1000,
		CacheTime:             30,
	}
	forwardConf := ServerConfig{
		UDPListenAddrs: []*net.UDPAddr{},
		TCPListenAddrs: []*net.TCPAddr{},
		FilteringConfig: FilteringConfig{
			ProtectionEnabled: true,
			UpstreamDNS:       []string{"8.8.8.8:53", "8.8.4.4:53"},
		},
		ConfigModified: func() {},
	}
	s := createTestServer(t, filterConf, forwardConf)

	defaultConf := s.conf

	err := s.Start()
	assert.Nil(t, err)
	t.Cleanup(func() {
		assert.Nil(t, s.Stop())
	})

	w := httptest.NewRecorder()

	testCases := []struct {
		name    string
		wantSet string
	}{{
		name:    "upstream_dns",
		wantSet: "",
	}, {
		name:    "bootstraps",
		wantSet: "",
	}, {
		name:    "blocking_mode_good",
		wantSet: "",
	}, {
		name:    "blocking_mode_bad",
		wantSet: "blocking_mode: incorrect value\n",
	}, {
		name:    "ratelimit",
		wantSet: "",
	}, {
		name:    "edns_cs_enabled",
		wantSet: "",
	}, {
		name:    "dnssec_enabled",
		wantSet: "",
	}, {
		name:    "cache_size",
		wantSet: "",
	}, {
		name:    "upstream_mode_parallel",
		wantSet: "",
	}, {
		name:    "upstream_mode_fastest_addr",
		wantSet: "",
	}, {
		name:    "upstream_dns_bad",
		wantSet: "wrong upstreams specification: missing port in address\n",
	}, {
		name:    "bootstraps_bad",
		wantSet: "a can not be used as bootstrap dns cause: invalid bootstrap server address: Resolver a is not eligible to be a bootstrap DNS server\n",
	}, {
		name:    "cache_bad_ttl",
		wantSet: "cache_ttl_min must be less or equal than cache_ttl_max\n",
	}, {
		name:    "upstream_mode_bad",
		wantSet: "upstream_mode: incorrect value\n",
	}, {
		name:    "local_ptr_upstreams_good",
		wantSet: "",
	}, {
		name:    "local_ptr_upstreams_null",
		wantSet: "",
	}}

	var data map[string]struct {
		Req  json.RawMessage `json:"req"`
		Want json.RawMessage `json:"want"`
	}
	loadTestData(t, t.Name()+jsonExt, &data)

	for _, tc := range testCases {
		caseData, ok := data[tc.name]
		require.True(t, ok)

		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				s.conf = defaultConf
			})

			rBody := ioutil.NopCloser(bytes.NewReader(caseData.Req))
			var r *http.Request
			r, err = http.NewRequest(http.MethodPost, "http://example.com", rBody)
			require.Nil(t, err)

			s.handleSetConfig(w, r)
			assert.Equal(t, tc.wantSet, w.Body.String())
			w.Body.Reset()

			s.handleGetConfig(w, nil)
			assert.JSONEq(t, string(caseData.Want), w.Body.String())
			w.Body.Reset()
		})
	}
}
