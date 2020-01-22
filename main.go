package main

import (
	gatewayv1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	kubegatewayv1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/apis/gateway.solo.io/v1"
	gwclientset "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/client/clientset/versioned"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/core/matchers"
	kubegloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/apis/gloo.solo.io/v1"
	glooclientset "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/client/clientset/versioned"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/options/static"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	skutils "github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"os"
)

func main() {

	// Read kubernetes configuration
	restCfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		os.Exit(1)
	}

	// Create a static upstream
	usMeta := core.Metadata{
		Name:      "postman-echo",
		Namespace: "gloo-system",
	}

	us := &kubegloov1.Upstream{
		ObjectMeta: skutils.ToKubeMeta(usMeta),
		Spec: gloov1.Upstream{
			Metadata: usMeta,
			UpstreamType: &gloov1.Upstream_Static{
				Static: &static.UpstreamSpec{
					Hosts: []*static.Host{
						{
							Addr: "postman-echo.com",
							Port: 80,
						},
					},
				},
			},
		},
	}

	// Get a client set for the Gloo API
	glooClientSet, err := glooclientset.NewForConfig(restCfg)
	if err != nil {
		os.Exit(1)
	}

	// Write the Upstream
	if _, err := glooClientSet.GlooV1().Upstreams("gloo-system").Create(us); err != nil {
		os.Exit(1)
	}

	// Create a virtual service pointing to the upstream
	vsMeta := core.Metadata{
		Name:      "hello",
		Namespace: "gloo-system",
	}

	vs := &kubegatewayv1.VirtualService{
		ObjectMeta: skutils.ToKubeMeta(vsMeta),
		Spec: gatewayv1.VirtualService{
			Metadata: vsMeta,
			VirtualHost: &gatewayv1.VirtualHost{
				Domains: []string{
					"*",
				},
				Routes: []*gatewayv1.Route{
					{
						Matchers: []*matchers.Matcher{
							{
								PathSpecifier: &matchers.Matcher_Prefix{Prefix: "/get"},
							},
						},
						Action: &gatewayv1.Route_RouteAction{
							RouteAction: &gloov1.RouteAction{
								Destination: &gloov1.RouteAction_Single{
									Single: &gloov1.Destination{
										DestinationType: &gloov1.Destination_Upstream{
											Upstream: &core.ResourceRef{
												Name:      "postman-echo",
												Namespace: "gloo-system",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Get a client set for the Gateway API
	gatewayClientSet, err := gwclientset.NewForConfig(restCfg)
	if err != nil {
		os.Exit(1)
	}

	// Write the Virtual Service
	if _, err = gatewayClientSet.GatewayV1().VirtualServices("gloo-system").Create(vs); err != nil {
		os.Exit(1)
	}
}
