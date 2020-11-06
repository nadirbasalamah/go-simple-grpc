package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/nadirbasalamah/go-simple-grpc/database"
	"github.com/nadirbasalamah/go-simple-grpc/model"
	"github.com/nadirbasalamah/go-simple-grpc/product/productpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
}

func (*server) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	return &productpb.CreateProductResponse{}, nil
}
func (*server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	return &productpb.GetProductResponse{}, nil
}
func (*server) EditProduct(ctx context.Context, req *productpb.EditProductRequest) (*productpb.EditProductResponse, error) {
	return &productpb.EditProductResponse{}, nil
}
func (*server) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {
	return &productpb.DeleteProductResponse{}, nil
}
func (*server) GetProducts(req *productpb.GetProductsRequest, stream productpb.ProductService_GetProductsServer) error {
	rows, err := database.DB.Query("SELECT id, name, description, category, amount FROM products ORDER BY name")
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	defer rows.Close()
	result := model.Products{}
	for rows.Next() {
		product := model.Product{}
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Category, &product.Amount)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Internal error: %v", err),
			)
		}
		result.Products = append(result.Products, product)
		stream.Send(&productpb.GetProductsResponse{
			Product: dataToProductPb(&product),
		})
	}

	if len(result.Products) == 0 {
		return status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Products not found: %v", err),
		)
	}

	return nil

}
func (*server) CreateBatchProduct(stream productpb.ProductService_CreateBatchProductServer) error {
	return nil
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
