package canvas

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// PublicAssets fetches public web assets only. Resolve and validate at dial time
// as well as on redirects, so a document cannot fetch host-local services.
func PublicAssets() FetchAsset {
	transport := &http.Transport{DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("DNS lookup failed")
		}
		for _, ip := range ips {
			if ip.IP.IsLoopback() || ip.IP.IsPrivate() || ip.IP.IsLinkLocalUnicast() || ip.IP.IsLinkLocalMulticast() || ip.IP.IsUnspecified() || ip.IP.IsMulticast() {
				return nil, fmt.Errorf("host-local resources must be imported as assets")
			}
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("no address for resource host")
		}
		return (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	}}
	client := &http.Client{Transport: transport, Timeout: 15 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		if req.URL.User != nil {
			return fmt.Errorf("resource credentials are not supported")
		}
		return nil
	}}
	return func(raw string) ([]byte, string, error) {
		resp, err := client.Get(raw)
		if err != nil {
			return nil, "", fmt.Errorf("resource could not be downloaded")
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("resource returned HTTP %d", resp.StatusCode)
		}
		data, err := ReadAsset(resp.Body)
		return data, resp.Header.Get("Content-Type"), err
	}
}
