# wiz
A versioned and signed database for large files

# status
- [ ] Created format specification, supported features
  - [x] b-tree like structures to increase overall performance e.g. due to cache locality and partial shadowing of nodes
  - [x] limitless snapshotting with constant effort for read and write. Deleting snapshots and rewriting history must be as efficient as possible.
  - [x] merkle tree to cryptographically sign commits and referenced data, so that wiz can be applied to blockchain products
  - [ ] support for not using merkle trees supporting performance sensitive use cases
  - [ ] support for deduplication using hash collisions together with byte comparision to guarantee correct storage even when affected by collision attacks
  - [ ] non-deduplication modes supporting performance / memory sensitive use cases
  - [ ] efficient support for cow transactions without accumulating snapshots (ever only one commit and at max one pending commit)
  - [ ] flash friendly reference counting for all nodes to care of deleted data for garbage collection
  - [ ] free space datastructure to efficiently find overwriteable pages
  - [ ] estimate compromise between system complexity and performance and choose wisely
  - [x] Transparent encryption
- [ ] Implement specification in go
- [ ] Create a command line application to open, inspect, repair, unpack and modify a wiz database
- [ ] Implement shared library for android, using go
- [ ] Implement a cross platform GUI (e.g. JavaFX) for the command line application

# TOC
* [wiz](#wiz)
* [status](#status)
* [Format specification](#format-specification)
  * [Nodes in general](#nodes-in-general)
  * [Header node](#header-node)
  * [Configuration node](#configuration-node)
  * [Boundary node](#boundary-node)
  * [Free node](#free-node)
  * [Reverse node](#reverse-node)
  * [Blob node](#blob-node)
  * [Data node](#data-node)
  * [Merkle data node](#merkle-data-node)
  * [Stream node](#stream-node)
  * [Commit node](#commit-node)
  * [Directory Node](#directory-node)
    * [Directory Node: Entry Node - File](#directory-node-entry-node---file)
    * [Directory Node: Entry Node - Directory](#directory-node-entry-node---directory)
    * [Directory Node: Entry Node - Overflow](#directory-node-entry-node---overflow)
  * [HIS-Node](#his-node)
    * [HT-Node](#ht-node)
    * [HO-Node](#ho-node)
  * [CR-Node](#cr-node)
    * [CL-Node](#cl-node)
  * [Encryption](#encryption)
    * [0x01: AES-256 CTR](#0x01-aes-256-ctr)

# Format specification
Basically wiz is just a specification of different nodes. In general a node should be kept so small, that the executing environment can process nodes in memory entirely (take also compression into account, especially after decompression), however it is up to the writer to decide which is best in the intended use case.

## Nodes in general
A node always starts with a byte identifier and is usually followed by a varuint so that it has usually a 2 byte overhead (the type and the length referring to 0-127 byte payload) which increases dynamically as the payload size increases. Overhead also depends on the used compression algorithm, if a node supports that at all. The payload of a node should not exceed something reasonable, e.g. ZFS uses 128KiB but you may even go into the range of MiB to increase efficiency. The UTF8 type is always prefixed with a varuint length to increase efficiency for short strings but allowing also more than the typical 64k bytes. All numbers are treated as big endian to match network byte order. Even if most operating systems are little endian today, one cannot ignore BE systems, so any code must be endian independent anyway. It is not expected that wiz may profit from a system specific endianess, like ZFS does.


## Header node
Marks a container and must be always the first node of a file and should not occur once again. If it does (e.g. for recovery purposes), it is not allowed to be contradictory. Wiz containers can simply be identified using the magic bytes [0x00 0x77 0x69 0x7a 0x63].
```
       name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x00         1 byte      byte
magic                  [77 69 7a 63]    4 byte      [4]byte
sub format magic         [* * * *]      4 byte      [4]byte
version                    0x03         4 byte      uint32
encryption                  *           1 byte      byte

```

| Name      | Value         | Length | Type    |
| --------- | :-----------: | -----: | ------: |
| node type | 0x00          | 1 byte | byte    |
| magic     | [77 69 7a 63] | 4 byte | [4]byte |

The version indicates which nodes and how they are defined. A node format may be changed in future revisions but should be extended in a backwards compatible manner. If such a thing is not possible (e.g. by adding new kinds of nodes) the number increases.

_Some notes to the version flag: Actually this is the third generation of the wiz format. The first only existed on paper, the second was implemented largely based on the paper based specification but is proprietary. So this is the first which is now open source. It is not only implemented in a different language but also does a lot of things differently to improve performance, storage usage, reliability and system complexity._

The known sub format identifiers of all known publicly available sub format identifiers.
```
wiza : the standard archive format of the command line tool
wizb : the format of the backup tool
```

The encryption formats are defined as follows:

```
0x00 : no encryption, all nodes are written as they are, just in plain text
0x01 : AES-256 CTR mode
```
See the encryption chapter for the detailed specification of each encryption mode.

## Configuration node
The wiz repository (as defined by the file) may include different properties. These properties are important to open the repository properly, e.g. picking the correct hash algorithm. This node must always be the second after the header.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x01         1 byte      byte
config version             0x00         4 byte      uint32
hash algorithm             0x*          1 byte      byte
pack id                     *           4 byte      uint32
length of key               *           1 byte      byte (e.g. for AES-256 this is 32)
encryption key              *           # bytes     []byte

```
The pack id cannot use the most significant bit, as it already indicates (in an 8bit pointer) if it is a local pointer offset or an external offset, so the maximum id is 2^31 = 2.147.483.648. In such references cases, cross pack pointers cannot be located beyond 4.294.967.296 byte. This allows to use pointers (not related to hash addressed nodes) to address 8 EiB (2^31*2^32) of storage.


The hash algorithms are used for merkle trees or similar structures and are defined as follows:
```
SHA256 : 0x00
```


## Boundary node
The writer is allowed to insert boundary nodes at will. Readers can use this to find back a synchronization point to recover nodes from damaged files. Nodes are not necessarily aligned to physical sectors. For example, if the length of a node has been deflected, all following nodes cannot be found trivially anymore. Using the boundary node allows a search algorithm to discard arbitrary bytes until it finds the next boundary. It is a tradeoff between wasting bytes for the boundary and setting the window for recovery which impacts loosing undamaged objects, because nodes cannot be read backwards (see also the Reverse node 0x04). However a writer may indeed insert free nodes (0x03) at will to ensure sector alignment of any kind and to ease recovery if this is a requirement.

```
     name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x02         1 byte      byte
the boundary             see below      [31]byte    byte

```

The boundary is defined as follows and is not intended to be changed. A scanner must be robust enough to detect random or attacking fake boundaries:
```
--wizBoundary7MA4YWxkTrZu0gW0gW
[2d 2d 77 69 7a 42 6f 75 6e 64 61 72 79 37 4d 41 34 59 57 78 6b 54 72 5a 75 30 67 57 30 67 57]
```

## Free node
Marks a free area. This may have been inserted after a delete operation or for padding purposes.

```
     name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x03         1 byte      byte
# free bytes                *          1-10 byte    varuint
free bytes                  *           # byte      []byte
```


## Reverse node
Nodes cannot be read backwards, because their header comes before data which is unspecified. Reading backwards is mostly not required at all, however due to the append only approach, the last written nodes are usually very interesting as an entry point.

```
      name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x04         1 byte      byte
offset bytes                *           4 byte      uint32
```

To read backwards, you have to read the last 1 + 4 bytes first. It must be of the form [0x04 * * * *]. Then seek back offset bytes + 5 bytes to start reading the previous node as always, starting at the node type identifier.



## Blob node
A blob node just contains payload bytes and has not a specific structure.


```
     name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x05         1 byte      byte
compression                 *           1 byte      byte
# data bytes                *          1-10 byte    varuint
data bytes                  *           # bytes     []byte
```
**adress = hash(0x05, uncompressed(data-bytes), length(uncompressed(bytes))**

The defined compressions are as follows:

```
NONE : 0x00
GZIP : 0x01
```

## Data node
A data node defines a list of 64 bit node pointers referring to either other data nodes or blob nodes. Evaluation needs to be recursive on data nodes to grab all blob nodes in the required order. Blob references must be in leaves. This allows the writer to create trees of blobs so that small changes within a large file and reusing large parts of the unchanged file becomes possible. The original wiz defined this only as an entire merkle tree, which may be unneeded at all, because integrity was ensured only at the commit and tree level anyway. If there is the requirement of a hash tree for data, use the merkle data node (0x07).

```
     name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x06         1 byte      byte
# of pointers               *          1-10 byte    varuint
64bit pointer bytes         *          # * 8 byte   []int64
```
**address = hash(0x06, #pointers, pointers...)**

## Merkle data node
It is like the simple data node (0x06) but referring to hashes instead of pointers. Blobs are also referenced in leaves. This can be used as an alternative to the stream node. You need to balance the additional 24 byte overhead per entry against the advantages of a hash tree.

```
      name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x07         1 byte      byte
# of hashes                 *          1-10 byte    varuint
hashes                      *         # * 32 byte   [][32]byte
```
**address = hash(0x07,#hashes, hashes...)**

## Stream node
When storing an actual stream of data, it's content is always represented by a root data node (0x06 or 0x07). The root node is referenced from the stream node which puts some more information on it like the actual size or the entire stream hash, which is stable, independent of the distribution of blobs (which is not possible in a hash tree) which comes into play for dynamic chunking. The stream hash is also calculated using a prefix and a postfix to avoid collision attacks.

```
      name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x08         1 byte      byte
stream length               *          1-10 byte    varuint
stream hash                 *           32 byte     [32]byte
offset                      *           8 byte      int64
```
**address = hash(0x08, length, hash)**

## Commit node
A commit incorporates a bunch of parent commits, a list of named trees, a message and a unix timestamp. The most important thing is that it does never contain pointers but the hashed values of the trees and parents. When calculating the hash of a node it should be always prefixed (e.g. type) and postfixed (e.g. length) as it is done by e.g. git and recommended by the Sakura hash tree mode to create a strong hash tree and to form a merkle tree.

Note that depending on the chosen data and stream nodes are not hashed directly and therefore are not part of the merkle tree. This is an explicit design decision to give writers the freedom to distribute data nodes at will to e.g. improve copy-on-write efficiency or e.g. remote delta updates by desired redundancy in different pack files.

```
      name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x09         1 byte      byte
timestamp                   *           8 byte      int64
message                     *           * byte      UTF8
# of parents                *           2 byte      uint16
hashes                      *          # * 32byte   [][32]byte
# of trees                  *           2 byte      uint16
(id  | hash)               *           # * byte    [](byte|[32]byte)
```
**address = hash(0x09, timestamp, message, #parents, hashes, #trees, ids|hashes...)**

The defined ids are as follows:

```
0x00 the root directory which represents a user navigatable file tree. Used by the archive and backup tool
0x01 the root tree for relational data
0x02 the root tree for the free space map (!? mixing a transaction logic with a commit)
0x03 the root tree for reference counting of nodes (!? mixing a transaction logic with a commit)
0x04 the root tree for the hash index tree (!? mixing a transaction logic with a commit)
```

## Directory Node
A directory node is a structured way of representing a logical (user defined) hierarchy. However as this may degenerate easily (e.g. thousands of entries per directory), the internal structure can be also an overflow node forming different kind of tree structures, e.g. a b-tree. The only specification which is set, is that the keys or overflow nodes are sorted ascending. Keys have to be unique.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x10         1 byte      byte
children hash               *           32 byte     [32]byte
# entries                   *           1-10byte    uvarint
entries                     *            *byte      []directory entry

```
**children hash = hash(sort(key)|entire entry-node|...)**  whereas only payload entry nodes and not overflow nodes (which do not have keys anyway) are considered

**address = hash(0x10, children-hash, #entries)**

The directory entries are defined as follows:

### Directory Node: Entry Node - File
A simple flat file entry node without any meta data.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
entry type                  0x11         1 byte      byte
# key bytes                  *          1-10 byte    varuint
key                          *           * byte      []byte
stream hash                  *           32 byte     [32]byte
```

### Directory Node: Entry Node - Directory
A simple directory entry node without any meta data.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
entry type                  0x12         1 byte      byte
# key bytes                  *          1-10 byte    varuint
key                          *           * byte      []byte
directory hash               *           32 byte     [32]byte
```

### Directory Node: Entry Node - Overflow
An overflow node refers to an offset pointer of another directory node (0x10) and which is not part of the cryptographic chain, because it just optimizes the data access (remember that the logical content for the hash tree is derived by the sorted list of the children). The referenced directory contains only values which are larger than any preceding keys and smaller than any succeeding keys (like in a b-tree). How the tree is actually organized is up-to the writer and the addressed use case. Because the referenced directory structure is a hash tree itself, an interesting effect is, that not only user defined defined trees can be deduplicated (still, overflow directory nodes are not directly part of the parent hash tree) but also the trees from overflows which allows (depending on the writer) to reuse even large parts of a changed "linear sorted list" from a single directory, containing potentionally millions of entries.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
entry type                  0x13         1 byte      byte
directory pointer            *           8 byte      int64
```

## HIS-Node
The hash-index-super-node is located either in the third node (after the header 0x00 and config 0x01) or is the second last node in a file, followed by the according reverse node 0x04. It's location varies due to the use case, e.g. an archive file which can only be created by writing to a stream without the possibility of random access would not be possible otherwise. However accessing the file where seeking is impossible or very expensive, may need the node in the first few bytes, e.g. when accessing a file from an online service, where the file can just be read from the beginning to the end. The super node points to a node within this or another file, as possible for pointers in general. The tree needs not to be signed, it is just used to optimize the access speed. Actually the entire storage concept could work without the hash index.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x14         1 byte      byte
pointer                     *           8 byte      int64
```

### HT-Node
A hash-tree-node can be located anywhere, however for better locality the writer is encouraged to put these nodes together, so that it can be read with a single page read. The HT-Node contains a list like the directory node (0x10) but does not need the third root node, because it's content need not to be signed. The writer can design to create a b-tree like structure to search and insert trees (the entries are always sorted by hash ascending). When appending nodes, the tree will likely be balanced. The writer may decide to overwrite existing nodes but may also just append the new ones to lower the risk of data corruption. A better approach could be to keep the index in a separate file to improve data locality and and avoid corrupting files containing important user data.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
node type                  0x15         1 byte      byte
# entries                   *           1-10byte    uvarint
entries                     *            *byte      []mixed array of HO- or HT-Node
```

### HO-Node
A simple hash-offset node contains the hash value and it's pointer location. It has a fixed size of 41 byte.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
entry type                 0x16         1 byte      byte
node hash                   *           32 byte     [32]byte
pointer                     *           8 byte      int64

```

## CR-Node
The content root node contains information which commits are the latest, regarding branches and tags. These are structured in hierarchical way. You can form a tag hierarchy just like folders, e.g. tags/v1 but you can create any structure like backup/2017-08-26 13:45 which fits your domain best. To save space, branches do not have a prefix like "branch/master", it is just "master". Entries are ordered ascending regarding their keys and can be interleaved with other CR-Nodes like in a b-tree. This allows the writer to optimize access times for the most important branches e.g. "master" and export uncommon parts like tags into overflowing CR-Nodes. This hierarchy is not signed and not part of the hash tree. It has to be the second last entry in the last added pack file of a pack set, followed by a reverse node (0x04). To improve data consistency a writer may decide to generally only add new data into a new file of the pack set, so that data is only written once and never overwritten. Losing the lastly added pack file will therefore always guarantee a valid pack set. When rewriting everything in a single pack file, the writer can also decide to append new CR-Roots to improve reliability. A truncated pack needs to be parsed from the beginning. The last valid parseable root node is then the required one.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
entry type                 0x17         1 byte      byte
# entries                   *           1-10byte    uvarint
entries                     *           * byte      []mixed array of CR- or CL-Node

```

### CL-Node
The content leaf node describes the name or key of a commit and it's hash.

```
       name                value        length         type
---------------------|---------------|-----------|----------------
entry type                 0x18         1 byte      byte
# key length                *           1-10byte    uvarint
commit hash                 *           32 byte     [32]byte
```

## Free space node
The free space node describes a tree for

## Reference counting node
The free space node describes a tree for

## Encryption
This chapter describes the encryption modes in detail. In general we do not need to use authenticated encryption because we already have an entire hash tree starting at a certain commit node (0x09), over directory nodes (0x10) pointing to either other directory nodes or file nodes which in turn references the stored stream data (e.g. 0x08). If required one can even store the contents of a stream in an independent hash tree (see merkle data node 0x06). Because of this design, an attacker may tamper the data, however this can be verified by the reader due to the hash tree nature. Tampering things which are not part the hash tree results in a recoverable corruption (e.g. when attacking a HT-Node)

### 0x01: AES-256 CTR
The CTR encryption mode is actually quite simple to handle. The AES block size is 16 byte, even for AES-256 which only means having a key which is 256 bit long. The IV must be as large as the block size, so 16 byte. Note that variants with a 256 bit block are not AES, but a form of RIJNDAEL in general. Important for CTR is that the IV just needs to be a nonce which is neither random nor secure. So the IV/nonce is defined to contain an incrementing 6 bytes unsigned number to identify the block (allowing 2^48 blocks which is enough to encrypt a file into the Exabyte range) and another 6 bytes as a re-write counter (probably exceeding a drives live). The remaining 4 bytes are assigned to the pack file nummer itself. This guarantees that no block of any pack file within a pack set ever gets the same IV, so we can guarantee the nonce requirement.

```
IV (16 byte) = <4 byte pack file id> | <6 byte write counter> | <6 byte block id>
```
To allow changing a password without a full rewrite of the encrypted data, a special treatment of the header (0x00) and config node (0x01) are required. The header node is never encrypted, so that a tool can always detect the (encrypted) format properly. However the config node (0x01) is always encrypted with the key which is derived from bcrypt value of the users secret password. The encrypted config node is always located at the block offset 2048 and contains the randomly generated key for all following blocks. The next node is located at block offset 4096 and encrypted with the key from the config node. Note that a blocks data can only contain 4080 bytes (16*255) with the prefixed nonce of 16 byte to get the fixed block size of 4096 bytes, so indeed the first actually encrypted node starts at offset 4096+16 bytes at the encrypted level.
