package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "os"
    "time"
)

func main() {
    url := os.Getenv("TARGET")
    if url == "" {
        panic("missing TARGET env variable")
    }

    dnsIP := os.Getenv("DNSIP")
    if url == "" {
        panic("missing DNSIP env variable")
    }

    for {
        resp, err := badHTTPClient(dnsIP).Get(url)
        if err != nil {
            log.Printf("ping failed: %v", err)
        } else {
            if resp.StatusCode == 200 {
                log.Println("ping succeeded")
            } else {
                payload, err := io.ReadAll(resp.Body)
                if err != nil {
                    log.Printf("an error occurred while reading the HTTP response: %v\n", err)
                } else {
                    resp.Body.Close()
                    log.Printf("ping failed: %s (code: %d)\n", string(payload), resp.StatusCode)
                }
            }
        }

        time.Sleep(2 * time.Second)
    }
}

func badHTTPClient(dnsIP string) *http.Client {
    dialer := &net.Dialer{
        Resolver: &net.Resolver{
            PreferGo: true,
            Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                d := net.Dialer{}
                fmt.Printf("dialing to %s\n", dnsIP)
                return d.DialContext(ctx, "udp", dnsIP)
            },
        },
    }

    dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
        return dialer.DialContext(ctx, network, addr)
    }

    http.DefaultTransport.(*http.Transport).DialContext = dialContext
    return &http.Client{}
}
