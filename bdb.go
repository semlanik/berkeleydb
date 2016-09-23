package berkeleydb

/*
#cgo LDFLAGS: -ldb
#include <db.h>
#include <stdlib.h>
#include "bdb.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

const BDB_VERSION string = "0.0.1"

// Flags for opening a database or environment.
const (
	DB_CREATE   = C.DB_CREATE
	DB_EXCL     = C.DB_EXCL
	DB_RDONLY   = C.DB_RDONLY
	DB_TRUNCATE = C.DB_TRUNCATE

	// Env
	DB_INIT_MPOOL = C.DB_INIT_MPOOL
)

// Database types.
const (
	DB_BTREE   = C.DB_BTREE
	DB_HASH    = C.DB_HASH
	DB_HEAP    = C.DB_HEAP
	DB_RECNO   = C.DB_RECNO
	DB_QUEUE   = C.DB_QUEUE
	DB_UNKNOWN = C.DB_UNKNOWN
)

type BDB struct {
	db *C.DB
}

type DBC struct {
	dbc *C.DBC
}

type Errno int

func NewDB() (*BDB, error) {
	var db *C.DB
	err := C.db_create(&db, nil, 0)

	if err > 0 {
		return nil, createError(err)
	}

	return &BDB{db}, nil
}

func NewDBInEnvironment(env *Environment) (*BDB, error) {
	var db *C.DB
	err := C.db_create(&db, env.environ, 0)

	if err > 0 {
		return nil, createError(err)
	}

	return &BDB{db}, nil
}

func (handle *BDB) OpenWithTxn(filename string, txn *C.DB_TXN, dbtype C.DBTYPE, flags C.u_int32_t) error {
	db := handle.db
	file := C.CString(filename)
	defer C.free(unsafe.Pointer(file))

	ret := C.go_db_open(db, txn, file, nil, dbtype, flags, 0)

	return createError(ret)
}

func (handle *BDB) Open(filename string, dbtype C.DBTYPE, flags C.u_int32_t) error {
	file := C.CString(filename)
	defer C.free(unsafe.Pointer(file))

	ret := C.go_db_open(handle.db, nil, file, nil, dbtype, flags, 0)

	return createError(ret)
}

func (handle *BDB) Close() error {
	ret := C.go_db_close(handle.db, 0)

	return createError(ret)
}

func (handle *BDB) OpenFlags() (C.u_int32_t, error) {
	var flags C.u_int32_t

	ret := C.go_db_get_open_flags(handle.db, &flags)

	return flags, createError(ret)
}

func (handle *BDB) Remove(filename string) error {
	file := C.CString(filename)
	defer C.free(unsafe.Pointer(file))

	ret := C.go_db_remove(handle.db, file)

	return createError(ret)
}

func (handle *BDB) Rename(oldname, newname string) error {
	oname := C.CString(oldname)
	defer C.free(unsafe.Pointer(oname))
	nname := C.CString(newname)
	defer C.free(unsafe.Pointer(nname))

	ret := C.go_db_rename(handle.db, oname, nname)

	return createError(ret)
}

// Convenience function to store a string.
func (handle *BDB) PutString(name, value string) error {
	nname := C.CString(name)
	defer C.free(unsafe.Pointer(nname))
	nvalue := C.CString(value)
	defer C.free(unsafe.Pointer(nvalue))

	ret := C.go_db_put_string(handle.db, nname, nvalue, 0)
	if ret > 0 {
		return createError(ret)
	}
	return nil
}

// Convenience function to get a string.
func (handle *BDB) GetString(name string) (string, error) {
	value := C.CString("")
	defer C.free(unsafe.Pointer(value))
	nname := C.CString(name)
	defer C.free(unsafe.Pointer(nname))

	ret := C.go_db_get_string(handle.db, nname, value)
	return C.GoString(value), createError(ret)
}

func (handle *BDB) DeleteString(name string) error {
	nname := C.CString(name)
	defer C.free(unsafe.Pointer(nname))

	ret := C.go_db_del_string(handle.db, nname)
	return createError(ret)
}

func (handle *BDB) Cursor() (*DBC, error) {
	var dbc *C.DBC

	err := C.go_db_cursor(handle.db, &dbc)

	if err > 0 {
		return nil, createError(err)
	}

	return &DBC{dbc}, nil
}

func (cursor *DBC) GetNext() (string, string, error) {
	value := C.CString("")
	defer C.free(unsafe.Pointer(value))
	key := C.CString("")
	defer C.free(unsafe.Pointer(key))

	ret := C.go_cursor_get_next(cursor.dbc, key, value)
	return C.GoString(key), C.GoString(value), createError(ret)
}

func (cursor *DBC) GetPrev() (string, string, error) {
	value := C.CString("")
	defer C.free(unsafe.Pointer(value))
	key := C.CString("")
	defer C.free(unsafe.Pointer(key))

	ret := C.go_cursor_get_prev(cursor.dbc, key, value)
	return C.GoString(key), C.GoString(value), createError(ret)
}

func (cursor *DBC) GetFirst() (string, string, error) {
	value := C.CString("")
	defer C.free(unsafe.Pointer(value))
	key := C.CString("")
	defer C.free(unsafe.Pointer(key))

	ret := C.go_cursor_get_first(cursor.dbc, key, value)
	return C.GoString(key), C.GoString(value), createError(ret)
}

func (cursor *DBC) GetLast() (string, string, error) {
	value := C.CString("")
	defer C.free(unsafe.Pointer(value))
	key := C.CString("")
	defer C.free(unsafe.Pointer(key))

	ret := C.go_cursor_get_last(cursor.dbc, key, value)
	return C.GoString(key), C.GoString(value), createError(ret)
}

// UTILITY FUNCTIONS

func Version() string {
	lib_version := C.GoString(C.db_full_version(nil, nil, nil, nil, nil))

	tpl := "%s (Go bindings v%s)"
	return fmt.Sprintf(tpl, lib_version, BDB_VERSION)
}

type DBError struct {
	Code    int
	Message string
}

func createError(code C.int) error {
	if code == 0 {
		return nil
	}
	msg := C.db_strerror(code)
	e := DBError{int(code), C.GoString(msg)}
	return errors.New(e.Error())
}

func (e *DBError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}
