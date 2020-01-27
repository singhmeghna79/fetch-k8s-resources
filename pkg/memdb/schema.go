package memdb

import (
	mdb "github.com/hashicorp/go-memdb"
	errors "github.com/pkg/errors"
	"strings"
)

// K8sResource ...
type K8sResource struct {
	Name            string
	Namespace       string
	Kind            string
	APIVersion      string
	Labels          map[string]string
	Annotations     map[string]string
	ResourceVersion string
	UID             string
	Data            map[string]interface{}
}

// Table ...
type Table struct {
	db   *mdb.MemDB
	name string
}

// Schema ...
type Schema struct {
	schema *mdb.DBSchema
	name   string
}

// NewSchemaForTable ...
func NewSchemaForTable(name string) *Schema {
	return &Schema{
		schema: &mdb.DBSchema{
			Tables: map[string]*mdb.TableSchema{
				name: &mdb.TableSchema{
					Name: name,
					Indexes: map[string]*mdb.IndexSchema{
						strings.ToLower(uid): &mdb.IndexSchema{
							Name:    strings.ToLower(uid),
							Unique:  true,
							Indexer: &mdb.StringFieldIndex{Field: uid},
						},
						strings.ToLower(name): &mdb.IndexSchema{
							Name:    strings.ToLower(name),
							Unique:  false,
							Indexer: &mdb.StringFieldIndex{Field: name},
						},
						strings.ToLower(namespace): &mdb.IndexSchema{
							Name:    strings.ToLower(namespace),
							Unique:  false,
							Indexer: &mdb.StringFieldIndex{Field: namespace},
						},
						strings.ToLower(kind): &mdb.IndexSchema{
							Name:    strings.ToLower(kind),
							Unique:  false,
							Indexer: &mdb.StringFieldIndex{Field: kind},
						},
						strings.ToLower(apiversion): &mdb.IndexSchema{
							Name:    strings.ToLower(apiversion),
							Unique:  false,
							Indexer: &mdb.StringFieldIndex{Field: apiversion},
						},
						strings.ToLower(resourceversion): &mdb.IndexSchema{
							Name:    strings.ToLower(resourceversion),
							Unique:  false,
							Indexer: &mdb.StringFieldIndex{Field: resourceversion},
						},
					},
				},
			},
		},
		name: name,
	}
}

// Apply ...
func (s *Schema) Apply() (*Table, error) {
	db, err := mdb.NewMemDB(s.schema)
	if err != nil {
		return nil, err
	}
	return &Table{
		db:   db,
		name: s.name,
	}, nil
}

// Save ...
func (t *Table) Save(r K8sResource) error {
	if t.name != r.Kind {
		return errors.New("kind mismatch")
	}
	txn := t.db.Txn(true)
	err := txn.Insert(t.name, r)
	if err != nil {
		txn.Abort()
		return err
	}
	txn.Commit()
	return nil
}
