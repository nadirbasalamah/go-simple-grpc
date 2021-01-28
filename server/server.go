package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"

	_ "github.com/lib/pq"
	"github.com/nadirbasalamah/go-simple-grpc/database"
	"github.com/nadirbasalamah/go-simple-grpc/model"
	"github.com/nadirbasalamah/go-simple-grpc/product/productpb"
	"github.com/nadirbasalamah/go-simple-grpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
}

func (*server) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	productReq := req.GetProduct()

	product := model.Product{
		Name:        productReq.GetName(),
		Description: productReq.GetDescription(),
		Category:    productReq.GetCategory(),
		Amount:      int(productReq.GetAmount()),
	}

	createdProduct, err := service.CreateProduct(product)
	if err != nil {
		return nil, err
	}

	return &productpb.CreateProductResponse{
		Product: dataToProductPb(&createdProduct),
	}, nil
}
func (*server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	id := req.GetProductId()

	product, err := service.GetProduct(id)
	if err != nil {
		return nil, err
	}

	return &productpb.GetProductResponse{
		Product: dataToProductPb(&product),
	}, nil
}
func (*server) EditProduct(ctx context.Context, req *productpb.EditProductRequest) (*productpb.EditProductResponse, error) {
	productReq := req.GetProduct()
	id := productReq.GetId()

	product := model.Product{
		Name:        productReq.GetName(),
		Description: productReq.GetDescription(),
		Category:    productReq.GetCategory(),
		Amount:      int(productReq.GetAmount()),
	}

	editedProduct, err := service.EditProduct(product, id)
	if err != nil {
		return nil, err
	}

	return &productpb.EditProductResponse{
		Product: dataToProductPb(&editedProduct),
	}, nil
}
func (*server) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {
	id := req.GetProductId()

	err := service.DeleteProduct(id)
	if err != nil {
		return nil, err
	}

	return &productpb.DeleteProductResponse{
		ProductId: id,
	}, nil
}
func (*server) GetProducts(req *productpb.GetProductsRequest, stream productpb.ProductService_GetProductsServer) error {
	err := service.GetProducts(stream)
	if err != nil {
		return err
	}
	return nil
}
func (*server) CreateBatchProduct(stream productpb.ProductService_CreateBatchProductServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&productpb.CreateBatchProductResponse{
				BatchResult: "All the data successfully inserted!",
			})
		}
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Internal error, insert batch failed: %v", err),
			)
		}
		product := model.Product{
			Name:        req.GetProduct().GetName(),
			Description: req.GetProduct().GetDescription(),
			Category:    req.GetProduct().GetCategory(),
			Amount:      int(req.GetProduct().GetAmount()),
		}

		_, err2 := database.DB.Query("INSERT INTO products (name, description, category, amount) VALUES ($1, $2, $3, $4) ", product.Name, product.Description, product.Category, product.Amount)
		if err2 != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Internal error, insert batch failed: %v", err),
			)
		}
	}
}

func dataToProductPb(data *model.Product) *productpb.Product {
	return &productpb.Product{
		Id:          int32(data.ID),
		Name:        data.Name,
		Description: data.Description,
		Category:    data.Category,
		Amount:      int32(data.Amount),
	}
}

func main() {
	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Product service started")

	// connect to DB
	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	// register product service server
	productpb.RegisterProductServiceServer(s, &server{})
	// enable gRPC reflection
	reflection.Register(s)

	go func() {
		fmt.Println("Starting server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until a signal is received
	<-ch
	fmt.Println("Stopping the server..")
	s.Stop()
	fmt.Println("Stopping listener...")
	lis.Close()
	fmt.Println("End of Program")
}
