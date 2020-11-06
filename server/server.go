package main

import (
	"context"
	"database/sql"
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

	_, err := database.DB.Query("INSERT INTO products (name, description, category, amount) VALUES ($1, $2, $3, $4) ", product.Name, product.Description, product.Category, product.Amount)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error, insert data failed: %v", err),
		)
	}

	return &productpb.CreateProductResponse{
		Product: dataToProductPb(&product),
	}, nil
}
func (*server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	id := req.GetProductId()
	product := model.Product{}

	row, err := database.DB.Query("SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error, data cannot be retrieved: %v", err),
		)
	}

	defer row.Close()

	for row.Next() {
		switch err := row.Scan(&product.ID, &product.Amount, &product.Name, &product.Description, &product.Category); err {
		case sql.ErrNoRows:
			return nil, status.Errorf(
				codes.NotFound,
				fmt.Sprintf("Data not found: %v", err),
			)
		case nil:
			log.Println(product.Name, product.Description, product.Category, product.Amount)
		default:
			return nil, status.Errorf(
				codes.Internal,
				fmt.Sprintf("Internal error, data cannot be retrieved: %v", err),
			)
		}
	}

	if product.ID == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Data not found: %v", err),
		)
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

	_, err := database.DB.Query("UPDATE products SET name=$1, description=$2, category=$3, amount=$4 WHERE id=$5", product.Name, product.Description, product.Category, product.Amount, id)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error, update data failed: %v", err),
		)
	}

	return &productpb.EditProductResponse{
		Product: dataToProductPb(&product),
	}, nil
}
func (*server) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {
	id := req.GetProductId()

	_, err := database.DB.Query("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error, delete data failed: %v", err),
		)
	}

	return &productpb.DeleteProductResponse{
		ProductId: id,
	}, nil
}
func (*server) GetProducts(req *productpb.GetProductsRequest, stream productpb.ProductService_GetProductsServer) error {
	rows, err := database.DB.Query("SELECT id, name, description, category, amount FROM products ORDER BY name")
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error, data cannot be retrieved: %v", err),
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
