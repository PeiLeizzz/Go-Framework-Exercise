## Go 标准库

### 输入输出

#### io：为 IO 原语提供基本的接口

> 不能假定这些操作是并发安全的。

- `Reader` 接口

    ```go
    type Reader interface {
      Read(p []byte) (n int, err error)
    }
    ```

- `Writer` 接口

    ```go
    type Writer interface {
      Write(p []byte) (n int, err error)
    }
    ```

- `ReaderAt` 接口

    ```go
    type ReaderAt interface {
      ReadAt(p []byte, off int64) (n int, err error)
    }
    ```

    - 当返回的 `n < len(p)` 时，会返回非 nil 错误（`Read` 接口不要求）
    - 若可读取的数据不到 `len(p)` 字节，`ReadAt` 就会阻塞，直到所有数据都可用或一个错误发生
    - 可对相同的输入源并发执行

- `WriterAt` 接口

    ```go
    type WriterAt interface {
      WriteAt(p []byte, off int64) (n int, err error)
    }
    ```

- `ReaderFrom` 接口

    ```go
    type ReaderFrom interface {
      ReadFrom(r Reader) (n int64, err error)
    }
    ```

    `Copy()` 函数会使用它（如果存在）

- `WriterTo` 接口

    ```go
    type WriterTo interface {
      WriteTo(w Writer) (n int64, err error)
    }
    ```

    `Copy()` 函数会使用它（如果存在）

- `Seeker` 接口

    ```go
    type Seeker interface {
      Seek(offset int64, whence int) (ret int64, err error)
    }
    ```

    > Seek 设置下一次 Read 或 Write 的偏移量为 oﬀset，它的解释取决于 whence：0 表示相对于文件的起始处，1 表示相对于当前的偏移，而 2 表示相对于其结尾处。Seek 返回新的偏移量和一个错误，如果有的话。

- `Closer` 接口

    ```go
    type Closer interface {
      Close() error
    }
    ```

- `ByteReader` 接口

    ```go
    type ByteReader interface {
      ReadByte() (c byte, err error)
    }
    ```

- `ByteWriter` 接口

    ```go
    type ByteWriter interface {
      WriteByte(c byte) error
    }
    ```

#### io/ioutil：封装一些实用的 I/O 函数

#### fmt：实现格式化 I/O

#### bufio：实现带缓冲 I/O