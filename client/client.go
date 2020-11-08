package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/nadirbasalamah/go-simple-grpc/product/productpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Client of product service")
	// create client server
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to product service: %v\n", err)
	}

	defer cc.Close()

	c := productpb.NewProductServiceClient(cc)

	// create a product
	createProduct(c)

	// get product by id
	getProductByID(c)

	// update a product
	updateProduct(c)

	// delete a product
	deleteProduct(c)

	// get all products
	getAllProducts(c)

	// create batch product (insert multiple products)
	createBatchProduct(c)
}

func createProduct(c productpb.ProductServiceClient) {
	fmt.Println("Create a product")
	req := &productpb.CreateProductRequest{
		Product: &productpb.Product{
			Name:        "Sample product",
			Category:    "Gadget",
			Amount:      int32(100),
			Description: "A sample product",
		},
	}

	res, err := c.CreateProduct(context.Background(), req)
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	fmt.Printf("Product created: %v\n", res)
}

func getProductByID(c productpb.ProductServiceClient) {
	fmt.Println("Get product data by ID")
	res, err := c.GetProduct(context.Background(), &productpb.GetProductRequest{
		ProductId: int32(19),
	})

	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	fmt.Printf("Product data: %v\n", res)
}

func updateProduct(c productpb.ProductServiceClient) {
	fmt.Println("Update a product")
	req := &productpb.EditProductRequest{
		Product: &productpb.Product{
			Id:          int32(9),
			Name:        "Sample Edited product",
			Category:    "Books",
			Amount:      int32(100),
			Description: "An edited product",
		},
	}

	res, err := c.EditProduct(context.Background(), req)
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", res)
	}

	fmt.Printf("Product updated: %v\n", res)
}

func deleteProduct(c productpb.ProductServiceClient) {
	fmt.Println("Delete a product")
	res, err := c.DeleteProduct(context.Background(), &productpb.DeleteProductRequest{
		ProductId: int32(26),
	})

	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	fmt.Printf("Product deleted: %v\n", res)
}

func getAllProducts(c productpb.ProductServiceClient) {
	fmt.Println("All products data")
	stream, err := c.GetProducts(context.Background(), &productpb.GetProductsRequest{})
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error when streaming: %v\n", err)
		}
		fmt.Println(res.GetProduct())
	}
}

func createBatchProduct(c productpb.ProductServiceClient) {
	requests := []*productpb.CreateBatchProductRequest{
		&productpb.CreateBatchProductRequest{
			Product: &productpb.Product{
				Name:        "Sample product one",
				Category:    "Gadget",
				Amount:      int32(100),
				Description: "A sample product one",
			},
		},
		&productpb.CreateBatchProductRequest{
			Product: &productpb.Product{
				Name:        "Sample product two",
				Category:    "Books",
				Amount:      int32(100),
				Description: "A sample product two",
			},
		},
		&productpb.CreateBatchProductRequest{
			Product: &productpb.Product{
				Name:        "Sample product three",
				Category:    "Foods",
				Amount:      int32(100),
				Description: "A sample product three",
			},
		},
	}

	stream, err := c.CreateBatchProduct(context.Background())
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	for _, req := range requests {
		fmt.Printf("Sending request: %v\n", req)
		stream.Send(req)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}
	fmt.Printf("Create batch product result: %v\n", res)
}
