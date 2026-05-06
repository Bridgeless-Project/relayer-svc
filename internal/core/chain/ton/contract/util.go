// Code generated; DO NOT EDIT.
package contract

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func stuckFileds[T Msg](data T) []fieldInfo {
	stack := make([]fieldInfo, 0)
	slices := make([]fieldInfo, 0)
	size := 0
	for i := range reflect.TypeOf(data).NumField() {
		v := reflect.TypeOf(data).Field(i)
		bit_size := v.Tag.Get("bit_size")
		bit_size_int, err := strconv.Atoi(bit_size)
		if err != nil {
			bit_size_int = -BitSize
		}
		if bit_size_int > 0 {
			size += bit_size_int
		}

		if v.Type.String() == "[]uint8" {
			slices = append(slices, fieldInfo{v.Type.String(), v.Name, bit_size_int, reflect.ValueOf(data).FieldByName(v.Name).Interface()})
			continue
		}

		switch v.Type.String() {
		case "address.Address":
			if size+AddressSize >= CellSize {
				stack = append(stack, slices...)
				slices = make([]fieldInfo, 0)
				size = 0
			}
			size += AddressSize
			stack = append(stack, fieldInfo{v.Type.String(), v.Name, bit_size_int, reflect.ValueOf(data).FieldByName(v.Name).Interface()})
			break
		case "bool":
			if size+BitSize >= CellSize {
				stack = append(stack, slices...)
				slices = make([]fieldInfo, 0)
				size = 0

			}
			size += BitSize
			stack = append(stack, fieldInfo{v.Type.String(), v.Name, bit_size_int, reflect.ValueOf(data).FieldByName(v.Name).Interface()})
			break
		case "*big.Int":
			intSize := 0
			if bit_size_int <= 0 {
				intSize = IntSize
			} else {
				intSize = bit_size_int
			}

			if size+intSize >= CellSize {
				stack = append(stack, slices...)
				slices = make([]fieldInfo, 0)
				size = 0

			}
			size += intSize
			stack = append(stack, fieldInfo{v.Type.String(), v.Name, bit_size_int, reflect.ValueOf(data).FieldByName(v.Name).Interface()})
			break
		}
	}

	return append(stack, slices...)
}

func constructCells(opcode uint64, stack []fieldInfo) ([]cell.Builder, error) {
	size := 0
	cells := make([]cell.Builder, 0)
	slice := cell.BeginCell()

	slice.StoreUInt(opcode, 32) // store op code
	for i, dataType := range stack {
		switch dataType.typeName {
		case "[]uint8": // parse Slice
			value, ok := dataType.value.([]byte)
			if !ok {
				return nil, fmt.Errorf("invalid address type")
			}
			subCell := cell.BeginCell()
			err := subCell.StoreSlice(value, uint(len(value))*8)
			if err != nil {
				return nil, err
			}

			err = slice.StoreRef(subCell.EndCell())
			if err != nil {
				return nil, err
			}
			break

		case "address.Address":
			if i != 0 && stack[i-1].typeName == "[]uint8" {
				cells = append(cells, *slice)
				size = 0
				slice = cell.BeginCell()

			}
			if size+AddressSize >= CellSize {

				newSlice := cell.BeginCell()
				value, ok := dataType.value.(address.Address)
				if !ok {
					return nil, fmt.Errorf("invalid address type")
				}
				if err := newSlice.StoreAddr(&value); err != nil {
					return nil, err
				}

				cells = append(cells, *slice)
				slice = newSlice

				size = AddressSize
				continue
			}
			value, ok := dataType.value.(address.Address)
			if !ok {
				return nil, fmt.Errorf("invalid address type")
			}
			err := slice.StoreAddr(&value)
			if err != nil {
				return nil, err
			}
			size += AddressSize

		case "bool":
			if i != 0 && stack[i-1].typeName == "[]uint8" {
				cells = append(cells, *slice)
				size = 0
				slice = cell.BeginCell()

			}
			if size+BitSize >= CellSize {

				value, ok := dataType.value.(bool)
				if !ok {
					return nil, fmt.Errorf("invalid boop type")
				}
				newSlice := cell.BeginCell()
				err := newSlice.StoreBoolBit(value)
				if err != nil {
					return nil, err
				}
				cells = append(cells, *slice.Copy())
				slice = newSlice

				size = BitSize
				continue
			}
			value, ok := dataType.value.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid bool type")
			}
			err := slice.StoreBoolBit(value)
			if err != nil {
				return nil, err
			}

			size += BitSize
		case "*big.Int":
			intSize := dataType.bitSize
			if intSize <= 0 {
				intSize = IntSize
			}
			if i != 0 && stack[i-1].typeName == "[]uint8" {
				cells = append(cells, *slice)
				size = 0
				slice = cell.BeginCell()
			}
			if size+intSize >= CellSize {
				newSlice := cell.BeginCell()
				value, ok := dataType.value.(*big.Int)
				if !ok {
					return nil, fmt.Errorf("invalid int type")
				}
				err := newSlice.StoreBigInt(value, uint(intSize))
				if err != nil {
					return nil, err
				}

				cells = append(cells, *slice.Copy())
				slice = newSlice
				size = intSize
				continue
			}

			value, ok := dataType.value.(*big.Int)
			if !ok {
				return nil, fmt.Errorf("invalid int type")
			}
			err := slice.StoreBigInt(value, uint(intSize))
			if err != nil {
				return nil, err
			}

			size += intSize
		}
	}

	if size > 0 {
		cells = append(cells, *slice)
	}

	return cells, nil
}

func aggregateCells(cells []cell.Builder) (*cell.Cell, error) {
	var resCell cell.Builder
	for i := len(cells) - 1; i >= 0; i-- {
		if i > 0 {
			err := cells[i-1].StoreRef(cells[i].EndCell())
			if err != nil {
				return nil, err
			}
		} else {
			resCell = cells[i]
		}
	}
	return resCell.EndCell(), nil
}

func BuildMessage[T Msg](opcode uint64, data T) (*cell.Cell, error) {
	stack := stuckFileds(data)
	cells, err := constructCells(opcode, stack)
	if err != nil {
		return nil, err
	}

	return aggregateCells(cells)
}

func parseLog[T Msg](boc cell.Cell, stack []fieldInfo, res *T) (*T, error) {
	bocSlice := boc.BeginParse()
	slice := bocSlice.Copy()
	var err error

	_, _ = slice.LoadInt(32) // skip op code
	for _, dataType := range stack {
		switch dataType.typeName {
		case "[]uint8": // parse Slice
			subSlice, err := slice.LoadRef()
			if err != nil {
				return nil, err
			}
			loadSlice, err := subSlice.LoadSlice(uint(len(subSlice.String())))
			if err != nil {
				return nil, err
			}

			reflect.ValueOf(res).Elem().FieldByName(dataType.fieldName).Set(reflect.ValueOf(loadSlice))
		case "address.Address":
			cellthis, _ := slice.ToCell()
			if cellthis.BitsSize() < AddressSize {
				slice, err = slice.LoadRef()
				if err != nil {
					return nil, err
				}
			}

			loadAddress, err := slice.LoadAddr()
			if err != nil {
				return nil, err
			}

			reflect.ValueOf(res).Elem().FieldByName(dataType.fieldName).Set(reflect.ValueOf(*loadAddress))
		case "bool":
			loadBool, err := slice.LoadBoolBit()
			if err != nil {
				return nil, err
			}

			reflect.ValueOf(res).Elem().FieldByName(dataType.fieldName).Set(reflect.ValueOf(loadBool))
		case "*big.Int":
			size := dataType.bitSize
			if size <= 0 {
				size = IntSize
			}
			cellthis, _ := slice.ToCell()
			if int(cellthis.BitsSize()) < size {
				slice, err = slice.LoadRef()
				if err != nil {
					return nil, err
				}

			}

			loadInt, err := slice.LoadBigInt(uint(size))
			if err != nil {
				return nil, err
			}
			reflect.ValueOf(res).Elem().FieldByName(dataType.fieldName).Set(reflect.ValueOf(loadInt))
		}
	}

	return res, nil
}

func ParseMessage[T Msg](boc cell.Cell, data T) (*T, error) {
	stack := stuckFileds(data)
	res, err := parseLog(boc, stack, &data)
	if err != nil {
		return nil, err
	}
	return res, nil
}
