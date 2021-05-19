package testing

import (
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	pb "github.com/topfreegames/podium/proto/podium/api/v1"
	"google.golang.org/grpc"
)

//SetupGRPC sets up the environment for grpc communication, starting the app and creating a connected client
func SetupGRPC(app *api.App, f func(pb.PodiumClient)) {
	InitializeTestServer(app)

	conn, err := grpc.Dial(app.GRPCEndpoint, grpc.WithInsecure())
	Expect(err).NotTo(HaveOccurred())
	defer func() {
		_ = conn.Close()
	}()

	cli := pb.NewPodiumClient(conn)

	f(cli)
}
