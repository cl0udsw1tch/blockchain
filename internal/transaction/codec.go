package transaction

import (
	"bytes"
	"encoding/binary"
)

const (
	OUTPOINT_TXID_SZ 	int = 32 	/ 8
	OUTPOINT_IDX_SZ  	int = 32 	/ 8
	TXOUT_VALUE_SZ		int = 64 	/ 8
)

type (
	BAD_OUTPOINT_ERR struct {}
	BAD_TXOUT_ERR struct {}
	BAD_TXIN_ERR struct {}
	BAD_TX_ERR struct {}
	BAD_SCRIPT_ERR struct {}
)

func (e BAD_OUTPOINT_ERR) Error() string {
	return "Incorrect outpoint format."
}
func (e BAD_TXOUT_ERR) Error() string {
	return "Incorrect txout format."
}

func (e BAD_TXIN_ERR) Error() string {
	return "Incorrect txin format."
}

func (e BAD_TX_ERR) Error() string {
	return "Incorrect tx format."
}

func (e BAD_SCRIPT_ERR) Error() string {
	return "Incorrect script format."
}

// outpoint codec ========================================= //

// **** decoder **** //
type OutPointDecoder struct {
	pt *OutPoint 
}

func (d *OutPointDecoder) New(pt *OutPoint){
	d.pt = pt
}

func (d *OutPointDecoder) Clear() {
	d.pt = nil
}

func (d *OutPointDecoder) Decode(buffer bytes.Buffer) error {


	if buffer.Len() != OUTPOINT_TXID_SZ + OUTPOINT_IDX_SZ {
		return BAD_OUTPOINT_ERR{}
	}
	b := buffer.Next(OUTPOINT_TXID_SZ + OUTPOINT_IDX_SZ)

	copy(d.pt.TxId, b[:32])
	d.pt.Idx = binary.BigEndian.Uint32(b[32:])


	return nil
}

func (d OutPointDecoder) Out() *OutPoint {
	return d.pt
}

// **** encoder **** //

type OutPointEncoder struct {
	buffer *bytes.Buffer
}

func (e *OutPointEncoder) New(buffer *bytes.Buffer){
	e.buffer = buffer
}

func (e *OutPointEncoder) Clear() {
	e.buffer = nil
}

func (e *OutPointEncoder) Encode(o OutPoint) {
	binary.Write(e.buffer, binary.BigEndian, o.TxId)
	binary.Write(e.buffer, binary.BigEndian, o.Idx)
}

func (e *OutPointEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// outpoint codec ========================================= //



// script codec ========================================= //

type ScriptBase struct {
	Size CompactSize
	Script [][]byte
}

// **** decoder **** //

type ScriptDecoder struct {
	script *ScriptBase
}

func (d *ScriptDecoder) New(s *ScriptBase) {
	d.script = s
}

func (d *ScriptDecoder) Clear() {
	d.script = nil
}

func (d *ScriptDecoder) Decode(buffer bytes.Buffer) error {
	if buffer.Len() < 1 {
		return BAD_SCRIPT_ERR{}
	}
	len := buffer.Next(1)[0]
	if buffer.Len() < int(len) {
		return BAD_SCRIPT_ERR{}
	}
	szValBytes := buffer.Next(int(len))
	szVal := binary.BigEndian.Uint32(szValBytes)
	d.script.Size = CompactSize{
		Type: len,
		Size: szValBytes,
	}
	if buffer.Len() < int(szVal) {
		return BAD_SCRIPT_ERR{}
	}
	var _byte byte
	var n uint32 = 0
	for n < uint32(szVal) {
		_byte = buffer.Next(1)[0]
		key := OpCode(_byte)
		if _, ok := OpMap[key]; ok {
			d.script.Script = append(d.script.Script, []byte{_byte})
			n += 1
		} else {
			switch _byte {
				case byte(OP_PUSHDATA1):
					d.script.Script = append(d.script.Script, []byte{byte(OP_PUSHDATA1)})
					if buffer.Len() < 1 {
						return BAD_SCRIPT_ERR{}
					}
					sz_bytes := buffer.Next(1)
					sz := sz_bytes[0]
					if buffer.Len() < int(sz) {
						return BAD_SCRIPT_ERR{}
					}
					d.script.Script = append(d.script.Script, sz_bytes, buffer.Next(int(sz)))
					n += 1 + 1 + uint32(sz)

				case byte(OP_PUSHDATA2):
					d.script.Script = append(d.script.Script, []byte{byte(OP_PUSHDATA2)})
					if buffer.Len() < 2 {
						return BAD_SCRIPT_ERR{}
					}
					sz_bytes := buffer.Next(2)
					sz := binary.BigEndian.Uint16(sz_bytes)
					if buffer.Len() < int(sz) {
						return BAD_SCRIPT_ERR{}
					}
					d.script.Script = append(d.script.Script, sz_bytes, buffer.Next(int(sz)))
					n += 1 + 2 + uint32(sz)

				case byte(OP_PUSHDATA4):
					d.script.Script = append(d.script.Script, []byte{byte(OP_PUSHDATA4)})
					if buffer.Len() < 4 {
						return BAD_SCRIPT_ERR{}
					}
					sz_bytes := buffer.Next(4)
					sz := binary.BigEndian.Uint32(sz_bytes)
					if buffer.Len() < int(sz) {
						return BAD_SCRIPT_ERR{}
					}
					d.script.Script = append(d.script.Script, sz_bytes, buffer.Next(int(sz)))
					n += 1 + 4 + sz
					
				default : 
					return BAD_SCRIPT_ERR{}
			}
		}
	}

	return nil
}

func (d *ScriptDecoder) Out() *ScriptBase {
	return d.script
}

// **** encoder **** //

type ScriptEncoder struct {
	buffer *bytes.Buffer
}

func (e ScriptEncoder) New(buffer *bytes.Buffer){
	e.buffer = buffer
}

func (e ScriptEncoder) Clear(){
	e.buffer = nil
}

func (e ScriptEncoder) Encode(s ScriptBase){
	
	binary.Write(e.buffer, binary.BigEndian, s.Size.Type)
	binary.Write(e.buffer, binary.BigEndian, s.Size.Size)

	for _, a := range s.Script {
		binary.Write(e.buffer, binary.BigEndian, a)
	}
}

func (e ScriptEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// script codec ========================================= //


// TxOut Codec ========================================= //

// **** decoder **** //

type TxOutDecoder struct {
	txout *TxOut 
}

func (d TxOutDecoder) New(txout *TxOut){
	d.txout = txout
}

func (d TxOutDecoder) Clear() {
	d.txout = nil
}

func (d TxOutDecoder) Decode(buffer bytes.Buffer) error {

	if buffer.Len() < TXOUT_VALUE_SZ {
		return BAD_TXOUT_ERR{}
	}
	d.txout.Value = int64(binary.BigEndian.Uint64(buffer.Next(TXOUT_VALUE_SZ)))

	script := ScriptBase{}
	var scriptD ScriptDecoder;
	scriptD.New(&script)
	err := scriptD.Decode(buffer)

	d.txout.LockingScriptSize = script.Size
	d.txout.LockingScript = script.Script

	return err	
}

func (d TxOutDecoder) Out() *TxOut {
	return d.txout
}

// **** encoder **** //

type TxOutEncoder struct {
	buffer *bytes.Buffer
}

func (e TxOutEncoder) New(buffer *bytes.Buffer){
	e.buffer = buffer
}

func (e TxOutEncoder) Clear(){
		e.buffer = nil
}

func (e TxOutEncoder) Encode(txOut TxOut){

	binary.Write(e.buffer, binary.BigEndian, txOut.Value)

	scriptEncoder := ScriptEncoder{}
	scriptEncoder.New(e.buffer)

	scriptBase := ScriptBase{
		Size: txOut.LockingScriptSize,
		Script: txOut.LockingScript,
	}
	scriptEncoder.Encode(scriptBase)

}

func (e TxOutEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// TxOut Codec ========================================= //

// TxIn Codec ========================================= //

// **** decoder **** //
type TxInDecoder struct {
	txin *TxIn
}

func (d *TxInDecoder) New(txin *TxIn) {
	d.txin = txin
}

func (d *TxInDecoder) Clear(){
	d.txin = nil
}

func (d TxInDecoder) Decode(buffer bytes.Buffer) error {
	outPointDecoder := OutPointDecoder{}
	outPointDecoder.New(&d.txin.PrevOutpt)
	err := outPointDecoder.Decode(buffer)
	if err != nil {
		return BAD_TXIN_ERR{}
	}
	scriptDecoder := ScriptDecoder{}
	err = scriptDecoder.Decode(buffer)
	if err != nil {
		return BAD_TXIN_ERR{}
	}

	script := scriptDecoder.Out()
	d.txin.UnlockingScriptSize = script.Size
	d.txin.UnlockingScript = script.Script

	return nil
}

func (d *TxInDecoder) Out() *TxIn{
	return d.txin
}

// **** encoder **** //

type TxInEncoder struct {
	buffer *bytes.Buffer
}

func (e *TxInEncoder) New(buffer *bytes.Buffer){
	e.buffer = buffer
}

func (e *TxInEncoder) Clear(){
	e.buffer = nil
}

func (e *TxInEncoder) Encode(txin TxIn){
	outPointEncoder := OutPointEncoder{}
	outPointEncoder.New(e.buffer)
	outPointEncoder.Encode(txin.PrevOutpt)

	scriptEncoder := ScriptEncoder{}
	scriptEncoder.New(e.buffer)
	scriptEncoder.Encode(ScriptBase{Size: CompactSize{Type: txin.UnlockingScriptSize.Type, Size: txin.UnlockingScriptSize.Size}})

}

func (e *TxInEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// TxIn Codec ========================================= //


// Tx Codec ========================================= //
// **** decoder **** //

type TxDecoder struct {
	tx *Tx
}

func (d *TxDecoder) New(tx *Tx){
	d.tx = tx
}

func (d *TxDecoder) Clear(){
	d.tx = nil
}

func (d *TxDecoder) Decode(buffer bytes.Buffer) error {
	if buffer.Len() < 4 {
		return BAD_TX_ERR{}
	}
	d.tx.Version = binary.BigEndian.Uint32(buffer.Next(4))

	if buffer.Len() < 1 {
		return BAD_TX_ERR{}
	}

	d.tx.NumInputs = buffer.Next(1)[0]

	for in := range d.tx.NumInputs {
		txinDecoder := TxInDecoder{}
		txinDecoder.New(&d.tx.Inputs[in])
		if err := txinDecoder.Decode(buffer); err != nil{
			return err
		}
		d.tx.Inputs[in] = *txinDecoder.Out()
	}

	if buffer.Len() < 1 {
		return BAD_TX_ERR{}
	}

	d.tx.NumOutputs = buffer.Next(1)[0]

	for out := range d.tx.NumOutputs {
		txoutDecoder := TxOutDecoder{}
		txoutDecoder.New(&d.tx.Outputs[out])
		if err := txoutDecoder.Decode(buffer); err != nil{
			return err
		}
		d.tx.Outputs[out] = *txoutDecoder.Out()
	}

	if buffer.Len() < 4 {
		return BAD_TX_ERR{}
	}
	d.tx.LockTime = binary.BigEndian.Uint32(buffer.Next(4)) 
	return nil
}

func (d *TxDecoder) Out() *Tx {
	return d.tx
}


// **** encoder **** //
type TxEncoder struct {
	buffer *bytes.Buffer
}

func (e *TxEncoder) New(buffer *bytes.Buffer){
	e.buffer = buffer
}

func (e *TxEncoder) Clear(){
	e.buffer = nil
}

func (e *TxEncoder) Encode(tx Tx){
	binary.Write(e.buffer, binary.BigEndian, tx.Version)
	binary.Write(e.buffer, binary.BigEndian, tx.NumInputs)
	for _, in := range tx.Inputs{
		inputEncoder := TxInEncoder{}
		inputEncoder.New(e.buffer)
		inputEncoder.Encode(in)
	}
	binary.Write(e.buffer, binary.BigEndian, tx.NumOutputs)
	for _, out := range tx.Outputs{
		outputEncoder := TxOutEncoder{}
		outputEncoder.New(e.buffer)
		outputEncoder.Encode(out)
	}
	binary.Write(e.buffer, binary.BigEndian, tx.LockTime)

}

func (e *TxEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}


// Tx Codec ========================================= //



