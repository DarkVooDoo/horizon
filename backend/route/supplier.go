package route

import (
	"encoding/json"
	"errors"
	"momo/db"
	"net/http"
)

func SupplierRoute(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		suppliers, err := Supplier{}.GetSuppliers()
		payload, jsonError := json.Marshal(suppliers)
		if err != nil || jsonError != nil {
			http.Error(res, err.Error(), http.StatusForbidden)
			return
		}
		res.Header().Add("Content-Type", "application/json")
		res.Write(payload)
	}
}

func (s Supplier) GetSuppliers() ([]Supplier, error) {
	var supplierName, supplierId string
	var supplier []Supplier = make([]Supplier, 0)
	suppliers, supplierErro := db.GetDbInstance().Query("SELECT supplier_id, supplier_name FROM Supplier")
	if supplierErro != nil {
		return nil, errors.New("supplier query failed")
	}
	defer suppliers.Close()
	for suppliers.Next() {
		suppliers.Scan(&supplierId, &supplierName)
		supplier = append(supplier, Supplier{supplierId, supplierName})
	}
	return supplier, nil
}

func (s Supplier) CreateSupplier() (string, error) {
	var supplierId string
	err := db.GetDbInstance().QueryRow("INSERT INTO Supplier (supplier_name) VALUES($1) RETURNING supplier_id", s.Name).Scan(&supplierId)
	if err != nil {
		return "", errors.New("supplier insert error")
	} else {
		return supplierId, nil
	}
}

type Supplier struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
