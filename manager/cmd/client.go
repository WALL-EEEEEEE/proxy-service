package cmd

import (
	"github.com/spf13/cobra"
)

var ClientCmd = &cobra.Command{
	Use:   "client",
	Short: "client for manager service",
	Run: func(cmd *cobra.Command, args []string) {

		/*
			var grpc_addr = ":8082"

			// Set up a connection to the server.
				conn, err := grpc.Dial(grpc_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("did not connect: %v", err)
				}
				defer conn.Close()
				c := pb.NewProxyServiceClient(conn)

				// Contact the server and print out its response.
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				r, err := c.Get(ctx, &pb.ProxyServiceGetRequest{})
				if err != nil {
					log.Fatalf("could not call service: %v", err)
				}
				log.Printf("Query Proxy: %s, %s, %+v", r.Status.GetStatus(), r.Status.GetMessage(), r.GetProxyList())
		*/
	},
}
