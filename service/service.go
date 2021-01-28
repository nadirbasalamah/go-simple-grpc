package service

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/nadirbasalamah/go-simple-grpc/database"
	"github.com/nadirbasalamah/go-simple-grpc/model"
	"github.com/nadirbasalamah/go-simple-grpc/product/productpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateProduct returns created product data
func CreateProduct(product model.Product) (model.Product, error) {
	_, err := database.DB.Query("INSERT INTO products (name, description, category, amount) VALUES ($1, $2, $3, $4) ", product.Name, product.Description, product.Category, product.Amount)
	if err != nil {
		return model.Product{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error, insert data failed: %v", err),
		)
	}
	return product, nil
}

// GetProduct returns specific product by id
func GetProduct(id int32) (model.Product, error) {
	product := model.Product{}

	row, err := database.DB.Query("SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		return model.Product{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error, data cannot be retrieved: %v", err),
		)
	}

	defer row.Close()

	for row.Next() {
		switch err := row.Scan(&product.ID, &product.Amount, &product.Name, &product.Description, &product.Category); err {
		case sql.ErrNoRows:
			return model.Product{}, status.Errorf(
				codes.NotFound,
				fmt.Sprintf("Data not found: %v", err),
			)
		case nil:
			log.Println(product.Name, product.Description, product.Category, product.Amount)
		default:
			return model.Product{}, status.Errorf(
				codes.Internal,
				fmt.Sprintf("Internal error, data cannot be retrieved: %v", err),
			)
		}
	}

	if product.ID == 0 {
		return model.Product{}, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Data not found: %v", err),
		)
	}

	return product, nil
}

// EditProduct returns edited product data
func EditProduct(product model.Product, id int32) (model.Product, error) {
	_, err := database.DB.Query("UPDATE products SET name=$1, description=$2, category=$3, amount=$4 WHERE id=$5", product.Name, product.Description, product.Category, product.Amount, id)
	if err != nil {
		return model.Product{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error, update data failed: %v", err),
		)
	}
	return product, nil
}

// DeleteProduct returns error occured when deleting a product data
func DeleteProduct(id int32) error {
	_, err := database.DB.Query("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error, delete data failed: %v", err),
		)
	}
	return nil
}

// GetProducts returns all product data
func GetProducts(stream productpb.ProductService_GetProductsServer) error {
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

func dataToProductPb(data *model.Product) *productpb.Product {
	return &productpb.Product{
		Id:          int32(data.ID),
		Name:        data.Name,
		Description: data.Description,
		Category:    data.Category,
		Amount:      int32(data.Amount),
	}
}
