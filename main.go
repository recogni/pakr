package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

////////////////////////////////////////////////////////////////////////////////

const (
	cPakMagic        = 0x5A6F12E1
	cPakFooterLength = 44
)

var (
	ErrUnsupportedVersion = errors.New("version not supported")
	ErrInvalidFooter      = errors.New("invalid footer data")
	ErrInvalidIndexSha1   = errors.New("invalid index record sha1 hash")
	ErrInvalidIndexRecord = errors.New("invalid index record")
)

////////////////////////////////////////////////////////////////////////////////

func grabStringN(r *bytes.Reader, n int) (string, error) {
	if r.Len() < n {
		return "", errors.New("buffer out of space")
	}

	ret := ""
	for i := 0; i < n; i++ {
		c, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if c == 0 {
			break
		}
		ret += string(c)
	}

	return ret, nil
}

////////////////////////////////////////////////////////////////////////////////

type Footer struct {
	magic   uint32   // constant 0x5A6F12E1
	version uint32   // 1, 2, or 3 - we only support v3
	offset  uint64   // offset of the index record
	size    uint64   // size of the index record
	hash    [20]byte // sha1 hash of the index record
}

func (f *Footer) Unmarshal(buf *bytes.Reader) error {
	binary.Read(buf, binary.LittleEndian, &f.magic)
	binary.Read(buf, binary.LittleEndian, &f.version)
	binary.Read(buf, binary.LittleEndian, &f.offset)
	binary.Read(buf, binary.LittleEndian, &f.size)

	var err error
	for i := 0; i < len(f.hash) && err == nil; i++ {
		f.hash[i], err = buf.ReadByte()
	}
	return err
}

////////////////////////////////////////////////////////////////////////////////

type Index struct {
	mountPointSize uint32
	mountPoint     string
	recordCount    uint32
	records        []*IndexRecord
}

func (idx *Index) Unmarshal(buf *bytes.Reader) error {
	var err error

	binary.Read(buf, binary.LittleEndian, &idx.mountPointSize)
	mpSz := int(idx.mountPointSize)
	if mpSz == 0 {
		return errors.New("invalid mount point size")
	}
	if buf.Len() < mpSz {
		return ErrInvalidIndexRecord
	}

	idx.mountPoint, err = grabStringN(buf, mpSz)
	if err != nil {
		return err
	}

	binary.Read(buf, binary.LittleEndian, &idx.recordCount)
	rc := int(idx.recordCount)
	for i := 0; i < rc; i++ {
		ir := &IndexRecord{}
		if err = ir.Unmarshal(buf); err != nil {
			return err
		}
		idx.records = append(idx.records, ir)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

type IndexRecord struct {
	fileNameSize uint32
	fileName     string
	metadata     *Record
}

func (ir *IndexRecord) Unmarshal(buf *bytes.Reader) error {
	var err error
	if err = binary.Read(buf, binary.LittleEndian, &ir.fileNameSize); err != nil {
		return err
	}
	ir.fileName, err = grabStringN(buf, int(ir.fileNameSize))
	if err != nil {
		return err
	}
	fmt.Printf(".. %s\n", ir.fileName)

	ir.metadata = &Record{}
	return ir.metadata.Unmarshal(buf)
}

////////////////////////////////////////////////////////////////////////////////

type Record struct {
	offset               uint64
	size                 uint64
	uncompressedSize     uint64
	compressionType      uint32
	hash                 [20]byte
	isEncrypted          byte
	compressionBlockSize uint32
}

func (r *Record) Unmarshal(buf *bytes.Reader) error {
	var err error
	if err = binary.Read(buf, binary.LittleEndian, &r.offset); err != nil {
		return err
	}
	if err = binary.Read(buf, binary.LittleEndian, &r.size); err != nil {
		return err
	}
	if err = binary.Read(buf, binary.LittleEndian, &r.uncompressedSize); err != nil {
		return err
	}
	if err = binary.Read(buf, binary.LittleEndian, &r.compressionType); err != nil {
		return err
	}
	for i := 0; i < len(r.hash) && err == nil; i++ {
		r.hash[i], err = buf.ReadByte()
	}
	if r.isEncrypted, err = buf.ReadByte(); err != nil {
		return err
	}
	if err = binary.Read(buf, binary.LittleEndian, &r.compressionBlockSize); err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func fatalOnError(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

func sha1Check(l, r [20]byte) error {
	for i := 0; i < 20; i++ {
		if l[i] != r[i] {
			return errors.New("sha1 hash mismatch")
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	pakFile := os.Args[1]
	bs, err := ioutil.ReadFile(pakFile)
	fatalOnError(err)

	// Step 1 :: Read the footer (last 44 bytes) and find the pointer
	//           to the index record.
	n := len(bs)
	if n <= cPakFooterLength {
		fatalOnError(ErrInvalidFooter)
	}

	buf := bytes.NewReader(bs)
	buf.Seek(int64(len(bs)-cPakFooterLength), io.SeekStart)

	footer := &Footer{}
	err = footer.Unmarshal(buf)
	fatalOnError(err)

	fmt.Printf("Pak file size: %d\n", n)
	fmt.Printf("index offset:  %d\n", footer.offset)
	fmt.Printf("index size:    %d\n", footer.size)

	// Step 2 :: Read the index record.
	if n < int(footer.offset+footer.size) {
		fatalOnError(ErrInvalidIndexRecord)
	}
	// Verify Sha1 sum for index record.
	ibs := bs[footer.offset : footer.offset+footer.size]
	if err = sha1Check(sha1.Sum(ibs), footer.hash); err != nil {
		fatalOnError(errors.New("sha1 mismatch for index record"))
	}
	// Unmarshal record from offset.
	buf.Seek(int64(footer.offset), io.SeekStart)
	index := &Index{}
	err = index.Unmarshal(buf)
	fatalOnError(err)

	// Step 3 :: Walk each record.
	for _, r := range index.records {
		fmt.Printf("File: %s :: %#v\n", r.fileName, r.metadata)
		buf.Seek(int64(r.metadata.offset), io.SeekStart)
		data := make([]byte, r.metadata.size)
		n, err := buf.ReadAt(data, int64(r.metadata.offset))
		fatalOnError(err)
		fmt.Printf("N  == %d\n", n)

		fmt.Printf("%#v\n", sha1.Sum(data))
		fmt.Printf("%#v\n", r.metadata.hash)
		if err = sha1Check(r.metadata.hash, sha1.Sum(data)); err != nil {
			fatalOnError(errors.New("sha1 mismatch for data record"))
		}
	}
}
