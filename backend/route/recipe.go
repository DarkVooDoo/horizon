package route

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"momo/db"
	"momo/util"
	"net/http"
	"strconv"

	"github.com/jung-kurt/gofpdf"
)

const (
	Horizontal_Margin     = 10
	Ingredient_Box_Height = 15
	Recipe_Box_Height     = 30
)

func RecipeRoute(res http.ResponseWriter, req *http.Request) {
	util.EnableCors(res, "http://localhost:3000")
	// res.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	var recipe FullRecipe
	var recipes Recipes
	if req.Method == http.MethodGet {
		if req.URL.Query().Has("rId") {
			rId := req.URL.Query().Get("rId")
			theRecipe, err := recipe.GetRecipe(rId)
			if err != nil {
				http.Error(res, err.Error(), http.StatusNoContent)
				return
			}
			payload, _ := json.Marshal(theRecipe)
			res.Header().Add("Content-Type", "application/json")
			res.Write(payload)
		} else if req.Header.Get("x-pdf") != "" {
			pdfId := req.Header.Get("x-pdf")
			r, _ := recipe.GetRecipe(pdfId)
			r.CreatePdfRecipe()
			res.Header().Set("Content-Disposition", `attachment; filename="test.pdf"`)
			http.ServeFile(res, req, "route/test.pdf")
		} else {
			var produit Produit
			produits, produitError := produit.GetProduits()
			myRecipes, err := recipes.GetRecipes()
			if err != nil || produitError != nil {
				http.Error(res, "error fetching recipes", http.StatusForbidden)
				return
			}
			recipesAndProduits := RecipeIngredients{Recipe: myRecipes, Produit: produits}
			payload, _ := json.Marshal(recipesAndProduits)
			res.Header().Add("Content-Type", "application/json")
			res.Write(payload)
		}
	} else if req.Method == http.MethodPost {
		err := util.ReadJSONBody(req.Body, &recipe)
		recipeError := recipe.CreateRecipe()
		if err != nil || recipeError != nil {
			fmt.Println(err)
			http.Error(res, "invalid body", http.StatusForbidden)
			return
		}
		res.Write([]byte("Recette creé"))
	} else if req.Method == http.MethodDelete {
		type Payload struct {
			Id string `json:"id"`
		}
		var payload Payload
		body, _ := io.ReadAll(req.Body)
		json.Unmarshal(body, &payload)
		err := recipe.DeleteRecipe(payload.Id)
		if err != nil {
			http.Error(res, "Forbidden", http.StatusForbidden)
			return
		}
		res.Write([]byte("Recette Supprimer"))
	}
}

func (r FullRecipe) CreateRecipe() error {
	var recipeId string
	tx, _ := db.GetDbInstance().Begin()
	re := tx.QueryRow("INSERT INTO Recipe (recipe_name) VALUES($1) RETURNING recipe_id", r.Name)
	if re.Err() != nil {
		fmt.Println("Recipe Err: ", re.Err())
		tx.Rollback()
		return errors.New("query failed")
	}
	re.Scan(&recipeId)
	for _, ingredient := range r.Ingredients {
		rows, err := tx.Query("INSERT INTO Ingredient (ingredient_quantity, ingredient_recipe_id, ingredient_produit_id) VALUES($1,$2,$3)", ingredient.Quantity, recipeId, ingredient.Id)
		rows.Close()
		if err != nil {
			fmt.Println("Ingredient Error: ", err)
			tx.Rollback()
			return errors.New("query failed")
		}
	}
	tx.Commit()
	return nil
}

func (Recipes) GetRecipes() ([]Recipes, error) {
	var recipes = make([]Recipes, 0)
	var name, id string
	var items uint16
	rows, posgresErr := db.GetDbInstance().Query("SELECT recipe_id, recipe_name, COUNT(ingredient_id) AS recipe_items FROM Recipe LEFT JOIN Ingredient ON recipe_id=ingredient_recipe_id GROUP BY recipe_id")
	if posgresErr != nil {
		return make([]Recipes, 0), errors.New("error fetching recipes")
	}
	for rows.Next() {
		rows.Scan(&id, &name, &items)
		recipes = append(recipes, Recipes{id, name, items})
	}
	return recipes, nil
}

func (FullRecipe) GetRecipe(id string) (FullRecipe, error) {
	var recipe_id, recipe_name, produit_name, produit_id string
	var quantity uint32
	var ingredient_price float32
	var ingredients = make([]Ingredient, 0)
	query := `SELECT recipe_id, recipe_name, produit_id, produit_name, ingredient_quantity, (produit_price/produit_packaging)*ingredient_quantity AS price FROM Recipe 
	LEFT JOIN Ingredient ON recipe_id=ingredient_recipe_id 
	LEFT JOIN Produit ON produit_id=ingredient_produit_id 
	WHERE recipe_id=$1`
	rows, err := db.GetDbInstance().Query(query, id)
	if err != nil {
		return FullRecipe{}, errors.New("cette recette n'existe pas")
	}
	for rows.Next() {
		rows.Scan(&recipe_id, &recipe_name, &produit_id, &produit_name, &quantity, &ingredient_price)
		if produit_id != "" {
			ingredients = append(ingredients, Ingredient{produit_id, produit_name, quantity, float32(math.Floor(float64(ingredient_price*100)) / 100)})
		}
	}
	var price float32 = 0
	for _, ingre := range ingredients {
		price += ingre.Price
	}
	return FullRecipe{Name: recipe_name, Id: recipe_id, Price: float32(math.Floor(float64(price*100)) / 100), Ingredients: ingredients}, nil
}

func (FullRecipe) DeleteRecipe(id string) error {
	_, err := db.GetDbInstance().Query("DELETE FROM Recipe WHERE recipe_id=$1", id)
	return err
}

func (r FullRecipe) CreatePdfRecipe() {
	var fontSize float64 = 18
	pdf := gofpdf.New(gofpdf.OrientationPortrait, gofpdf.UnitMillimeter, gofpdf.PageSizeA4, "")
	pdf.AddPage()
	width, _ := pdf.GetPageSize()
	pdf.SetFont("Arial", "B", fontSize)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFillColor(20, 92, 158)
	pdf.CellFormat(width-(Horizontal_Margin*2), Recipe_Box_Height, r.Name, "1", 1, "C M", true, 0, "")
	fontSize = 12
	pdf.SetFontSize(fontSize)
	var ingredientCellWith float64 = (width - (Horizontal_Margin * 2)) / 3
	var headers = []string{"Produit", "Recette de base", strconv.Itoa(int(r.Variant)) + "X"}
	pdf.SetTextColor(0, 0, 0)
	for headerIndex, head := range headers {
		if headerIndex == 0 {
			pdf.SetFillColor(215, 208, 235)
		} else if headerIndex == 1 {
			pdf.SetFillColor(141, 179, 231)
		} else {
			pdf.SetFillColor(238, 179, 159)
		}
		if headerIndex < len(headers)-1 {
			pdf.CellFormat(ingredientCellWith, Ingredient_Box_Height, head, "1", 0, "C M", true, 0, "")
		} else {
			pdf.CellFormat(ingredientCellWith, Ingredient_Box_Height, head, "1", 1, "C M", true, 0, "")
		}
	}
	pdf.SetFillColor(255, 255, 255)
	for _, ingredient := range r.Ingredients {
		pdf.CellFormat(ingredientCellWith, Ingredient_Box_Height, ingredient.Name, "1", 0, "C M", false, 0, "")
		pdf.CellFormat(ingredientCellWith, Ingredient_Box_Height, strconv.Itoa(int(ingredient.Quantity)), "1", 0, "C M", false, 0, "")
		pdf.CellFormat(ingredientCellWith, Ingredient_Box_Height, strconv.Itoa(int(ingredient.Quantity*uint32(r.Variant))), "1", 1, "C M", false, 0, "")
	}

	err := pdf.OutputFileAndClose("route/test.pdf")
	if err != nil {
		log.SetPrefix("Error\t")
		log.Panicln(err)
	}
}

type FullRecipe struct {
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	Price       float32      `json:"price"`
	Variant     uint16       `json:"variant"`
	Ingredients []Ingredient `json:"ingredients"`
}

type Ingredient struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity uint32  `json:"quantity"`
	Price    float32 `json:"price"`
}

type Recipes struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Items uint16 `json:"items"`
}

type RecipeIngredients struct {
	Recipe  []Recipes `json:"recipe"`
	Produit []Produit `json:"produit"`
}
