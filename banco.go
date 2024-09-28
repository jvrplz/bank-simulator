package main

import (
	"fmt"
	"sync"
)

// Estructura de la cuenta
type Cuenta struct {
	ID    int
	Saldo float64
	Mutex sync.Mutex
}

// Banco representa un conjunto de cuentas y un canal para manejar las transacciones
type Banco struct {
	Cuentas       map[int]*Cuenta
	Transacciones chan Transaccion
	wg            sync.WaitGroup
}

// Transaccion representa una transacción en el banco
type Transaccion struct {
	Tipo      string
	OrigenID  int
	DestinoID int
	Monto     float64
}

// CrearBanco inicializa un banco con cuentas vacías
func CrearBanco() *Banco {
	return &Banco{
		Cuentas:       make(map[int]*Cuenta),
		Transacciones: make(chan Transaccion, 100),
	}
}

// AgregarCuenta añade una nueva cuenta al banco
func (b *Banco) AgregarCuenta(id int, saldoInicial float64) {
	b.Cuentas[id] = &Cuenta{ID: id, Saldo: saldoInicial}
}

// Depositar agrega un monto a la cuenta de destino
func (b *Banco) Depositar(id int, monto float64) {
	cuenta := b.Cuentas[id]
	cuenta.Mutex.Lock()
	defer cuenta.Mutex.Unlock()
	cuenta.Saldo += monto
}

// Retirar resta un monto de la cuenta de origen
func (b *Banco) Retirar(id int, monto float64) bool {
	cuenta := b.Cuentas[id]
	cuenta.Mutex.Lock()
	defer cuenta.Mutex.Unlock()
	if cuenta.Saldo >= monto {
		cuenta.Saldo -= monto
		return true
	}
	return false
}

// Transferir realiza una transferencia de fondos entre dos cuentas
func (b *Banco) Transferir(origenID, destinoID int, monto float64) {
	if b.Retirar(origenID, monto) {
		b.Depositar(destinoID, monto)
	} else {
		fmt.Printf("Transferencia fallida: Fondos insuficientes en la cuenta %d\n", origenID)
	}
}

// ProcesarTransacciones procesa las transacciones de la cola
func (b *Banco) ProcesarTransacciones() {
	defer b.wg.Done()
	for transaccion := range b.Transacciones {
		switch transaccion.Tipo {
		case "depositar":
			b.Depositar(transaccion.DestinoID, transaccion.Monto)
			fmt.Printf("Depositados %.2f en la cuenta %d\n", transaccion.Monto, transaccion.DestinoID)
		case "retirar":
			if b.Retirar(transaccion.OrigenID, transaccion.Monto) {
				fmt.Printf("Retirados %.2f de la cuenta %d\n", transaccion.Monto, transaccion.OrigenID)
			} else {
				fmt.Printf("Retiro fallido: Fondos insuficientes en la cuenta %d\n", transaccion.OrigenID)
			}
		case "transferir":
			b.Transferir(transaccion.OrigenID, transaccion.DestinoID, transaccion.Monto)
			fmt.Printf("Transferidos %.2f de la cuenta %d a la cuenta %d\n", transaccion.Monto, transaccion.OrigenID, transaccion.DestinoID)
		}
	}
}

// AgregarTransaccion agrega una nueva transacción al canal de transacciones
func (b *Banco) AgregarTransaccion(transaccion Transaccion) {
	b.Transacciones <- transaccion
}

// Iniciar comienza el procesamiento de transacciones
func (b *Banco) Iniciar() {
	b.wg.Add(1)
	go b.ProcesarTransacciones()
}

// Finalizar cierra el canal de transacciones y espera a que se completen
func (b *Banco) Finalizar() {
	close(b.Transacciones)
	b.wg.Wait()
}

// Menú interactivo
func mostrarMenu() {
	fmt.Println("\n=== Menú de Banco ===")
	fmt.Println("1. Depositar dinero")
	fmt.Println("2. Retirar dinero")
	fmt.Println("3. Transferir dinero")
	fmt.Println("4. Ver saldo de una cuenta")
	fmt.Println("5. Salir")
	fmt.Print("Seleccione una opción: ")
}

func main() {
	banco := CrearBanco()

	// Crear cuentas
	banco.AgregarCuenta(1, 1000)
	banco.AgregarCuenta(2, 500)
	banco.AgregarCuenta(3, 2000)

	// Iniciar procesamiento de transacciones
	banco.Iniciar()

	var opcion int

	for {
		mostrarMenu()
		fmt.Scanln(&opcion)

		switch opcion {
		case 1:
			var cuentaID int
			var monto float64
			fmt.Print("Ingrese el ID de la cuenta para depositar: ")
			fmt.Scanln(&cuentaID)
			fmt.Print("Ingrese el monto a depositar: ")
			fmt.Scanln(&monto)
			banco.AgregarTransaccion(Transaccion{Tipo: "depositar", DestinoID: cuentaID, Monto: monto})

		case 2:
			var cuentaID int
			var monto float64
			fmt.Print("Ingrese el ID de la cuenta para retirar: ")
			fmt.Scanln(&cuentaID)
			fmt.Print("Ingrese el monto a retirar: ")
			fmt.Scanln(&monto)
			banco.AgregarTransaccion(Transaccion{Tipo: "retirar", OrigenID: cuentaID, Monto: monto})

		case 3:
			var origenID, destinoID int
			var monto float64
			fmt.Print("Ingrese el ID de la cuenta de origen: ")
			fmt.Scanln(&origenID)
			fmt.Print("Ingrese el ID de la cuenta de destino: ")
			fmt.Scanln(&destinoID)
			fmt.Print("Ingrese el monto a transferir: ")
			fmt.Scanln(&monto)
			banco.AgregarTransaccion(Transaccion{Tipo: "transferir", OrigenID: origenID, DestinoID: destinoID, Monto: monto})

		case 4:
			var cuentaID int
			fmt.Print("Ingrese el ID de la cuenta para ver el saldo: ")
			fmt.Scanln(&cuentaID)
			cuenta, existe := banco.Cuentas[cuentaID]
			if existe {
				fmt.Printf("Saldo de la cuenta %d: %.2f\n", cuentaID, cuenta.Saldo)
			} else {
				fmt.Printf("La cuenta con ID %d no existe.\n", cuentaID)
			}

		case 5:
			fmt.Println("Saliendo del programa...")
			banco.Finalizar()
			return

		default:
			fmt.Println("Opción no válida. Inténtelo de nuevo.")
		}
	}
}
