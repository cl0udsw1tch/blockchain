package transaction

import (
	"bytes"
	"encoding/binary"
)

const (
	OUTPOINT_TXID_SZ int32 = 32
	OUTPOINT_IDX_SZ  int32 = 4
	TXOUT_VALUE_SZ   int32 = 8
)

type (
	BAD_OUTPOINT_ERR struct{}
	BAD_TXOUT_ERR    struct{}
	BAD_TXIN_ERR     struct{}
	BAD_TX_ERR       struct{}
	BAD_SCRIPT_ERR   struct{}
	BAD_UTXO_ERR     struct{}
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

func (e BAD_UTXO_ERR) Error() string {
	return "Incorrect UTXO format."
}

// outpoint codec ========================================= //

// **** decoder **** //
type OutPointDecoder struct {
	pt *OutPoint
}

func NewOutPointDecoder(pt *OutPoint) *OutPointDecoder {
	dec := new(OutPointDecoder)
	if pt == nil {
		dec.pt = new(OutPoint)
	} else {
		dec.pt = pt
	}
	return dec
}

func (d *OutPointDecoder) Clear() {
	d.pt = new(OutPoint)
}

func (d *OutPointDecoder) Decode(buffer *bytes.Buffer) error {

	if int32(buffer.Len()) != OUTPOINT_TXID_SZ+OUTPOINT_IDX_SZ {
		return BAD_OUTPOINT_ERR{}
	}
	b := buffer.Next(int(OUTPOINT_TXID_SZ + OUTPOINT_IDX_SZ))

	copy(d.pt.TxId, b[:32])
	d.pt.Idx = int32(binary.BigEndian.Uint32(b[32:]))

	return nil
}

func (d OutPointDecoder) Out() *OutPoint {
	return d.pt
}

// **** encoder **** //

type OutPointEncoder struct {
	buffer *bytes.Buffer
}

func NewOutPointEncoder(buffer *bytes.Buffer) *OutPointEncoder {
	enc := new(OutPointEncoder)
	if buffer == nil {	
		enc.buffer = new(bytes.Buffer)
	} else {
		enc.buffer = buffer
	}
	return enc
}

func (e *OutPointEncoder) Clear() {
	e.buffer = new(bytes.Buffer)
}

func (e *OutPointEncoder) Encode(o *OutPoint) {
	binary.Write(e.buffer, binary.BigEndian, o.TxId)
	binary.Write(e.buffer, binary.BigEndian, o.Idx)
}

func (e *OutPointEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// outpoint codec ========================================= //

// script codec ========================================= //

type ScriptBase struct {
	Size   CompactSize
	Script [][]byte
}

// **** decoder **** //

type ScriptDecoder struct {
	script *ScriptBase
}

func NewScriptDecoder(script *ScriptBase) *ScriptDecoder {
	dec := new(ScriptDecoder)
	if script == nil {
		dec.script = new(ScriptBase)
	} else {
		dec.script = script
	}
	return dec
}

func (dec *ScriptDecoder) Clear() {
	dec.script = new(ScriptBase)
}

func (dec *ScriptDecoder) Decode(buffer *bytes.Buffer) error {
	if buffer.Len() < 1 {
		return BAD_SCRIPT_ERR{}
	}
	len := buffer.Next(1)[0]
	if buffer.Len() < int(len) {
		return BAD_SCRIPT_ERR{}
	}
	szValBytes := buffer.Next(int(len))
	szVal := binary.BigEndian.Uint32(szValBytes)
	dec.script.Size = CompactSize{
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
			dec.script.Script = append(dec.script.Script, []byte{_byte})
			n += 1
		} else {
			switch _byte {
			case byte(OP_PUSHDATA1):
				dec.script.Script = append(dec.script.Script, []byte{byte(OP_PUSHDATA1)})
				if buffer.Len() < 1 {
					return BAD_SCRIPT_ERR{}
				}
				sz_bytes := buffer.Next(1)
				sz := sz_bytes[0]
				if buffer.Len() < int(sz) {
					return BAD_SCRIPT_ERR{}
				}
				dec.script.Script = append(dec.script.Script, sz_bytes, buffer.Next(int(sz)))
				n += 1 + 1 + uint32(sz)

			case byte(OP_PUSHDATA2):
				dec.script.Script = append(dec.script.Script, []byte{byte(OP_PUSHDATA2)})
				if buffer.Len() < 2 {
					return BAD_SCRIPT_ERR{}
				}
				sz_bytes := buffer.Next(2)
				sz := binary.BigEndian.Uint16(sz_bytes)
				if buffer.Len() < int(sz) {
					return BAD_SCRIPT_ERR{}
				}
				dec.script.Script = append(dec.script.Script, sz_bytes, buffer.Next(int(sz)))
				n += 1 + 2 + uint32(sz)

			case byte(OP_PUSHDATA4):
				dec.script.Script = append(dec.script.Script, []byte{byte(OP_PUSHDATA4)})
				if buffer.Len() < 4 {
					return BAD_SCRIPT_ERR{}
				}
				sz_bytes := buffer.Next(4)
				sz := binary.BigEndian.Uint32(sz_bytes)
				if buffer.Len() < int(sz) {
					return BAD_SCRIPT_ERR{}
				}
				dec.script.Script = append(dec.script.Script, sz_bytes, buffer.Next(int(sz)))
				n += 1 + 4 + sz

			default:
				return BAD_SCRIPT_ERR{}
			}
		}
	}

	return nil
}

func (dec *ScriptDecoder) Out() *ScriptBase {
	return dec.script
}

// **** encoder **** //

type ScriptEncoder struct {
	buffer *bytes.Buffer
}

func NewScriptEncoder(buffer *bytes.Buffer) *ScriptEncoder {
	enc := new(ScriptEncoder)
	if buffer == nil {	
		enc.buffer = new(bytes.Buffer)
	} else {
		enc.buffer = buffer
	}
	return enc
}

func (e *ScriptEncoder) Clear() {
	e.buffer = new(bytes.Buffer)
}

func (e *ScriptEncoder) Encode(s *ScriptBase) {

	binary.Write(e.buffer, binary.BigEndian, s.Size.Type)
	binary.Write(e.buffer, binary.BigEndian, s.Size.Size)

	for _, a := range s.Script {
		binary.Write(e.buffer, binary.BigEndian, a)
	}
}

func (e *ScriptEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// script codec ========================================= //

// TxOut Codec ========================================= //

// **** decoder **** //

type TxOutDecoder struct {
	txout *TxOut
}

func NewTxOutDecoder(txout *TxOut) *TxOutDecoder {
	dec := new(TxOutDecoder)
	if txout == nil {
		dec.txout = new(TxOut)
	} else {
		dec.txout = txout
	}
	return dec
}

func (d *TxOutDecoder) Clear() {
	d.txout = new(TxOut)
}

func (d *TxOutDecoder) Decode(buffer *bytes.Buffer) error {

	if int32(buffer.Len()) < TXOUT_VALUE_SZ {
		return BAD_TXOUT_ERR{}
	}
	d.txout.Value = int64(binary.BigEndian.Uint64(buffer.Next(int(TXOUT_VALUE_SZ))))

	script := ScriptBase{}
	scriptD := NewScriptDecoder(&script)
	err := scriptD.Decode(buffer)

	d.txout.LockingScriptSize = script.Size
	d.txout.LockingScript = script.Script

	return err
}

func (d *TxOutDecoder) Out() *TxOut {
	return d.txout
}

// **** encoder **** //

type TxOutEncoder struct {
	buffer *bytes.Buffer
}

func NewTxOutEncoder(buffer *bytes.Buffer) *TxOutEncoder {
	enc := new(TxOutEncoder)
	if buffer == nil {	
		enc.buffer = new(bytes.Buffer)
	} else {
		enc.buffer = buffer
	}
	return enc
}

func (e *TxOutEncoder) Clear() {
	e.buffer = new(bytes.Buffer)
}

func (e *TxOutEncoder) Encode(txOut *TxOut) {

	binary.Write(e.buffer, binary.BigEndian, txOut.Value)

	scriptEncoder := NewScriptEncoder(e.buffer)

	scriptBase := ScriptBase{
		Size:   txOut.LockingScriptSize,
		Script: txOut.LockingScript,
	}
	scriptEncoder.Encode(&scriptBase)

}

func (e *TxOutEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// TxOut Codec ========================================= //

// TxIn Codec ========================================= //

// **** decoder **** //
type TxInDecoder struct {
	txin *TxIn
}

func NewTxInDecoder(txin *TxIn) *TxInDecoder {
	dec := new(TxInDecoder)
	if txin == nil {
		dec.txin = new(TxIn)
	} else {
		dec.txin = txin
	}
	return dec
}

func (d *TxInDecoder) Clear() {
	d.txin = new(TxIn)
}

func (d TxInDecoder) Decode(buffer *bytes.Buffer) error {

	outPointDecoder := NewOutPointDecoder(&d.txin.PrevOutpt)
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

func (d *TxInDecoder) Out() *TxIn {
	return d.txin
}

// **** encoder **** //

type TxInEncoder struct {
	buffer *bytes.Buffer
}

func NewTxInEncoder(buffer *bytes.Buffer) *TxInEncoder {
	enc := new(TxInEncoder)
	if buffer == nil {	
		enc.buffer = new(bytes.Buffer)
	} else {
		enc.buffer = buffer
	}
	return enc
}

func (e *TxInEncoder) Clear() {
	e.buffer = new(bytes.Buffer)
}

func (e *TxInEncoder) Encode(txin *TxIn) {

	outPointEncoder := NewOutPointEncoder(e.buffer)
	outPointEncoder.Encode(&txin.PrevOutpt)

	scriptEncoder := NewScriptEncoder(e.buffer)
	scriptEncoder.Encode(&ScriptBase{Size: CompactSize{Type: txin.UnlockingScriptSize.Type, Size: txin.UnlockingScriptSize.Size}})

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

func NewTxDecoder(tx *Tx) *TxDecoder {
	dec := new(TxDecoder)
	if tx == nil {
		dec.tx = new(Tx)
	} else {
		dec.tx = tx
	}
	return dec
}

func (d *TxDecoder) Clear() {
	d.tx = new(Tx)
}

func (d *TxDecoder) Decode(buffer *bytes.Buffer) error {
	if buffer.Len() < 4 {
		return BAD_TX_ERR{}
	}
	d.tx.Version = int32(binary.BigEndian.Uint32(buffer.Next(4)))

	if buffer.Len() < 1 {
		return BAD_TX_ERR{}
	}

	d.tx.NumInputs = buffer.Next(1)[0]

	for in := range d.tx.NumInputs {
		txinDecoder := NewTxInDecoder(&d.tx.Inputs[in])
		if err := txinDecoder.Decode(buffer); err != nil {
			return err
		}
		d.tx.Inputs[in] = *txinDecoder.Out()
	}

	if buffer.Len() < 1 {
		return BAD_TX_ERR{}
	}

	d.tx.NumOutputs = buffer.Next(1)[0]

	for out := range d.tx.NumOutputs {
		txoutDecoder := NewTxOutDecoder(&d.tx.Outputs[out])
		if err := txoutDecoder.Decode(buffer); err != nil {
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

func NewTxEncoder(buffer *bytes.Buffer) *TxEncoder {
	enc := new(TxEncoder)
	if buffer == nil {	
		enc.buffer = new(bytes.Buffer)
	} else {
		enc.buffer = buffer
	}
	return enc
}

func (e *TxEncoder) Clear() {
	e.buffer = new(bytes.Buffer)
}

func (e *TxEncoder) Encode(tx *Tx) {
	binary.Write(e.buffer, binary.BigEndian, tx.Version)
	binary.Write(e.buffer, binary.BigEndian, tx.NumInputs)
	for _, in := range tx.Inputs {

		inputEncoder := NewTxInEncoder(e.buffer)
		inputEncoder.Encode(&in)
	}
	binary.Write(e.buffer, binary.BigEndian, tx.NumOutputs)
	for _, out := range tx.Outputs {
		outputEncoder := NewTxOutEncoder(e.buffer)
		outputEncoder.Encode(&out)
	}
	binary.Write(e.buffer, binary.BigEndian, tx.LockTime)

}

func (e *TxEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// Tx Codec ========================================= //

// UTXO Codec ======================================== //
// **** decoder **** //
type UtxoDecoder struct {
	utxo *Utxo
}

func NewUtxoDecoder(utxo *Utxo) *UtxoDecoder {
	dec := new(UtxoDecoder)
	if utxo == nil {
		dec.utxo = new(Utxo)
	} else {
		dec.utxo = utxo
	}
	return dec
}

func (d *UtxoDecoder) Clear() {
	d.utxo = new(Utxo)
}

func (d *UtxoDecoder) Decode(buffer *bytes.Buffer) error {

	dec := NewOutPointDecoder(&d.utxo.OutPoint)
	if err := dec.Decode(buffer); err != nil {
		return BAD_UTXO_ERR{}
	}

	if buffer.Len() < 8 {
		return BAD_UTXO_ERR{}
	}

	d.utxo.Value = int64(binary.BigEndian.Uint64(buffer.Next(4)))

	scriptdec := NewScriptDecoder(&ScriptBase{Size: d.utxo.LockingScriptSize, Script: d.utxo.LockingScript})
	if err := scriptdec.Decode(buffer); err != nil {
		return BAD_UTXO_ERR{}
	}

	return nil
}

func (d UtxoDecoder) Out() *Utxo {
	return d.utxo
}

// **** encoder **** //

type UtxoEncoder struct {
	buffer *bytes.Buffer
}

func NewUtxoEncoder(buffer *bytes.Buffer) *UtxoEncoder {
	enc := new(UtxoEncoder)
	if buffer == nil {	
		enc.buffer = new(bytes.Buffer)
	} else {
		enc.buffer = buffer
	}
	return enc
}

func (e *UtxoEncoder) Clear() {
	e.buffer = new(bytes.Buffer)
}

func (e *UtxoEncoder) Encode(utxo *Utxo) {

	outEnc := NewOutPointEncoder(e.buffer)
	outEnc.Encode(&utxo.OutPoint)

	e.buffer.Write([]byte{byte(utxo.Value)})

	scriptenc := NewScriptEncoder(e.buffer)
	scriptenc.Encode(&ScriptBase{Size: utxo.LockingScriptSize, Script: utxo.LockingScript})

}

func (e *UtxoEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// UTXO Codec ======================================== //
