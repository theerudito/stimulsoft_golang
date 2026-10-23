package main

type Cabecera struct {
	Orden   string `json:"orden"`
	Cliente string `json:"cliente"`
	Piso    string `json:"piso"`
	Mesa    string `json:"mesa"`
	Mesero  string `json:"mesero"`
	Fecha   string `json:"fecha"`
	Total   string `json:"total"`
}

type Detalle struct {
	Cantidad    int    `json:"cantidad"`
	Descripcion string `json:"descripcion"`
}

type Pedido struct {
	Cabecera Cabecera  `json:"cabecera"`
	Detalles []Detalle `json:"detalles"`
}

var pedido = Pedido{
	Cabecera: Cabecera{
		Orden:   "106985",
		Cliente: "JORGE LOOR",
		Piso:    "1",
		Mesa:    "Mesa 2",
		Mesero:  "MESERO",
		Fecha:   "27/03/2026",
		Total:   "20.5",
	},
	Detalles: []Detalle{
		{Cantidad: 10, Descripcion: "CONTAFACIL PLAN 10 EMPRESAS"},
		{Cantidad: 20, Descripcion: "CONTAFACIL PLAN 20 EMPRESAS"},
		{Cantidad: 30, Descripcion: "CONTAFACIL PLAN 30 EMPRESAS"},
	},
}
