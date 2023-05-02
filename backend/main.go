package main

import (
	"momo/route"
	"net/http"
)

func main() {
	http.HandleFunc("/api/supplier", route.SupplierRoute)
	http.HandleFunc("/api/produit", route.ProduitRoute)
	http.HandleFunc("/api/recipe", route.RecipeRoute)
	http.ListenAndServe(":5000", nil)
}
