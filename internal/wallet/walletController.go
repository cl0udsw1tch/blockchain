package wallet

func CreateWallet(name string){
	CreateWalletFile(name)
}

func ReadWallet(name string) Wallet {
	return ReadWalletFile(name)
}

func ListWallets() []Wallet {
	return ReadWalletList()
}

func UpdateWallet(){

}

func DeleteWallet(){

}