package route

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"momo/db"
	"momo/util"
	"net/http"
)

func ProduitRoute(res http.ResponseWriter, req *http.Request) {
	util.EnableCors(res, "http://localhost:3000")
	var produit Produit
	if req.Method == http.MethodGet {
		produits, err := produit.GetProduits()
		payload, _ := json.Marshal(produits)
		if err != nil {
			fmt.Println(err)
		}
		res.Header().Add("Content-Type", "application/json")
		res.Write(payload)
	} else if req.Method == http.MethodPost {
		util.ReadJSONBody(req.Body, &produit)
		payload, err := produit.CreateProduit()
		if err != nil {
			http.Error(res, err.Error(), http.StatusForbidden)
		}
		res.Header().Add("Content-Type", "application/json")
		res.Write(payload)
	} else if req.Method == http.MethodDelete {
		type DeleteProduit struct {
			Id string `json:"id"`
		}
		var pId DeleteProduit
		err := util.ReadJSONBody(req.Body, &pId)
		dbError := produit.DeleteProduit(pId.Id)
		if err != nil || dbError != nil {
			http.Error(res, "forbidden", http.StatusForbidden)
			return
		}
		res.Write([]byte("Produit supprimer"))
	} else if req.Method == http.MethodPatch {
		util.ReadJSONBody(req.Body, &produit)
		updateProduit := produit.ModifyProduit()
		payload, _ := json.Marshal(updateProduit)
		res.Header().Add("Content-Type", "application/json")
		res.Write(payload)
	}
}

func (Produit) GetProduits() ([]Produit, error) {
	var produits []Produit = make([]Produit, 0)
	var name, id string
	var packaging uint
	var price float32
	const Query = "SELECT produit_id, produit_name, produit_packaging, produit_price FROM Produit"
	rows, err := db.GetDbInstance().Query(Query)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("query failed")
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&id, &name, &packaging, &price)
		produits = append(produits, Produit{Id: id, Name: name, Packaging: packaging, Price: price})
	}
	return produits, nil
}

func (p Produit) CreateProduit() ([]byte, error) {
	var id, name, supplier string
	var packaging uint
	var price float32
	var row *sql.Row
	if p.IsNewSupplier {
		mySupplier := Supplier{Name: p.Supplier}
		supplierId, _ := mySupplier.CreateSupplier()
		const query = "INSERT INTO Produit (produit_name, produit_packaging, produit_price, produit_supplier_id) VALUES($1,$2,ROUND($3,2), $4) RETURNING produit_id, produit_name, produit_packaging, produit_price"
		produitError := db.GetDbInstance().QueryRow(query, p.Name, p.Packaging, p.Price, supplierId).Scan(&id, &name, &packaging, &price)
		if produitError != nil {
			return nil, errors.New("error on insert")
		}

	} else {
		const Query = `INSERT INTO Produit (produit_name, produit_packaging, produit_price, produit_supplier_id) VALUES($1,$2,ROUND($3,2), $4) RETURNING produit_id, produit_name, produit_packaging, produit_price, produit_supplier_id`
		row = db.GetDbInstance().QueryRow(Query, p.Name, p.Packaging, p.Price, p.Supplier)
		row.Scan(&id, &name, &packaging, &price)
	}
	pr := Produit{Id: id, Name: name, Supplier: supplier, Packaging: packaging, Price: price}
	data, _ := json.Marshal(pr)
	return data, nil
}

func (Produit) DeleteProduit(id string) error {
	rows, err := db.GetDbInstance().Query("DELETE FROM Produit WHERE produit_id=$1", id)
	if err != nil {
		return errors.New("couldnt delete the produit")
	}
	defer rows.Close()
	return nil
}

func (p Produit) ModifyProduit() Produit {
	var id, name string
	var packaging uint
	var price float32
	const query = `UPDATE Produit SET produit_name=$1, produit_packaging=$2, produit_price=ROUND($3, 2) WHERE produit_id=$4
	RETURNING produit_id, produit_name, produit_packaging, produit_price`
	row := db.GetDbInstance().QueryRow(query, p.Name, p.Packaging, p.Price, p.Id)
	row.Scan(&id, &name, &packaging, &price)
	p = Produit{Id: id, Name: name, Packaging: packaging, Price: price}
	return p
}

type Produit struct {
	Id            string  `json:"id"`
	IsNewSupplier bool    `json:"newSupplier"`
	Name          string  `json:"name"`
	Supplier      string  `json:"supplier"`
	Packaging     uint    `json:"packaging"`
	Price         float32 `json:"price"`
}
