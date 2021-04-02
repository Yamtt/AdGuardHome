package dnsforward

import (
	"bufio"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsfilter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		name string
		conf func() ServerConfig
		want string
	}{{
		name: "all_right",
		conf: func() ServerConfig {
			return defaultConf
		},
		want: "{\"upstream_dns\":[\"8.8.8.8:53\",\"8.8.4.4:53\"],\"upstream_dns_file\":\"\",\"bootstrap_dns\":[\"9.9.9.10\",\"149.112.112.10\",\"2620:fe::10\",\"2620:fe::fe:10\"],\"protection_enabled\":true,\"ratelimit\":0,\"blocking_mode\":\"\",\"blocking_ipv4\":\"\",\"blocking_ipv6\":\"\",\"edns_cs_enabled\":false,\"dnssec_enabled\":false,\"disable_ipv6\":false,\"upstream_mode\":\"\",\"cache_size\":0,\"cache_ttl_min\":0,\"cache_ttl_max\":0,\"resolve_clients\":false,\"local_ptr_upstreams\":[]}\n",
	}, {
		name: "fastest_addr",
		conf: func() ServerConfig {
			conf := defaultConf
			conf.FastestAddr = true

			return conf
		},
		want: "{\"upstream_dns\":[\"8.8.8.8:53\",\"8.8.4.4:53\"],\"upstream_dns_file\":\"\",\"bootstrap_dns\":[\"9.9.9.10\",\"149.112.112.10\",\"2620:fe::10\",\"2620:fe::fe:10\"],\"protection_enabled\":true,\"ratelimit\":0,\"blocking_mode\":\"\",\"blocking_ipv4\":\"\",\"blocking_ipv6\":\"\",\"edns_cs_enabled\":false,\"dnssec_enabled\":false,\"disable_ipv6\":false,\"upstream_mode\":\"fastest_addr\",\"cache_size\":0,\"cache_ttl_min\":0,\"cache_ttl_max\":0,\"resolve_clients\":false,\"local_ptr_upstreams\":[]}\n",
	}, {
		name: "parallel",
		conf: func() ServerConfig {
			conf := defaultConf
			conf.AllServers = true

			return conf
		},
		want: "{\"upstream_dns\":[\"8.8.8.8:53\",\"8.8.4.4:53\"],\"upstream_dns_file\":\"\",\"bootstrap_dns\":[\"9.9.9.10\",\"149.112.112.10\",\"2620:fe::10\",\"2620:fe::fe:10\"],\"protection_enabled\":true,\"ratelimit\":0,\"blocking_mode\":\"\",\"blocking_ipv4\":\"\",\"blocking_ipv6\":\"\",\"edns_cs_enabled\":false,\"dnssec_enabled\":false,\"disable_ipv6\":false,\"upstream_mode\":\"parallel\",\"cache_size\":0,\"cache_ttl_min\":0,\"cache_ttl_max\":0,\"resolve_clients\":false,\"local_ptr_upstreams\":[]}\n",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(w.Body.Reset)

			s.conf = tc.conf()
			s.handleGetConfig(w, nil)

			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, tc.want, w.Body.String())
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
		req     string
		wantSet string
		wantGet string
	}{{
		name:    "upstream_dns",
		req:     "{\"upstream_dns\":[\"8.8.8.8:77\",\"8.8.4.4:77\"]}",
		wantSet: "",
	}, {
		name:    "bootstraps",
		req:     "{\"bootstrap_dns\":[\"9.9.9.10\"]}",
		wantSet: "",
	}, {
		name:    "blocking_mode_good",
		req:     "{\"blocking_mode\":\"refused\"}",
		wantSet: "",
	}, {
		name:    "blocking_mode_bad",
		req:     "{\"blocking_mode\":\"custom_ip\"}",
		wantSet: "blocking_mode: incorrect value\n",
	}, {
		name:    "ratelimit",
		req:     "{\"ratelimit\":6}",
		wantSet: "",
	}, {
		name:    "edns_cs_enabled",
		req:     "{\"edns_cs_enabled\":true}",
		wantSet: "",
	}, {
		name:    "dnssec_enabled",
		req:     "{\"dnssec_enabled\":true}",
		wantSet: "",
	}, {
		name:    "cache_size",
		req:     "{\"cache_size\":1024}",
		wantSet: "",
	}, {
		name:    "upstream_mode_parallel",
		req:     "{\"upstream_mode\":\"parallel\"}",
		wantSet: "",
	}, {
		name:    "upstream_mode_fastest_addr",
		req:     "{\"upstream_mode\":\"fastest_addr\"}",
		wantSet: "",
	}, {
		name:    "upstream_dns_bad",
		req:     "{\"upstream_dns\":[\"\"]}",
		wantSet: "wrong upstreams specification: missing port in address\n",
	}, {
		name:    "bootstraps_bad",
		req:     "{\"bootstrap_dns\":[\"a\"]}",
		wantSet: "a can not be used as bootstrap dns cause: invalid bootstrap server address: Resolver a is not eligible to be a bootstrap DNS server\n",
	}, {
		name:    "cache_bad_ttl",
		req:     "{\"cache_ttl_min\":1024,\"cache_ttl_max\":512}",
		wantSet: "cache_ttl_min must be less or equal than cache_ttl_max\n",
	}, {
		name:    "upstream_mode_bad",
		req:     "{\"upstream_mode\":\"somethingelse\"}",
		wantSet: "upstream_mode: incorrect value\n",
	}, {
		name:    "local_ptr_upstreams_good",
		req:     "{\"local_ptr_upstreams\":[\"123.123.123.123\"]}",
		wantSet: "",
	}, {
		name:    "local_ptr_upstreams_null",
		req:     "{\"local_ptr_upstreams\":null}",
		wantSet: "",
	}}

	var f *os.File
	f, err = os.Open("testdata/handleSetConfig.txt")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	sc := bufio.NewScanner(f)
	for i := range testCases {
		require.True(t, sc.Scan())

		testCases[i].wantGet = sc.Text() + "\n"
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				s.conf = defaultConf
			})

			rBody := ioutil.NopCloser(strings.NewReader(tc.req))
			var r *http.Request
			r, err = http.NewRequest(http.MethodPost, "http://example.com", rBody)
			require.Nil(t, err)

			s.handleSetConfig(w, r)
			assert.Equal(t, tc.wantSet, w.Body.String())
			w.Body.Reset()

			s.handleGetConfig(w, nil)
			assert.Equal(t, tc.wantGet, w.Body.String())
			w.Body.Reset()
		})
	}
}
