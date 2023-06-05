package bond

import (
	"context"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/go-bond/bond/serializers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBondTable_UnsafeUpdate(t *testing.T) {
	db := setupDatabase()
	defer tearDownDatabase(db)

	const (
		TokenBalanceTableID = TableID(1)
	)

	tokenBalanceTable := NewTable[TokenBalance](TableOptions[TokenBalance]{
		DB:        db,
		TableID:   TokenBalanceTableID,
		TableName: "token_balance",
		TablePrimaryKeyFunc: func(builder KeyBuilder, tb TokenBalance) []byte {
			return builder.AddUint64Field(tb.ID).Bytes()
		},
		Serializer: &serializers.JsonSerializer[TokenBalance]{},
	})

	tokenBalanceAccount := TokenBalance{
		ID:              1,
		AccountID:       1,
		ContractAddress: "0xtestContract",
		AccountAddress:  "0xtestAccount",
		Balance:         5,
	}

	tokenBalanceAccountUpdated := TokenBalance{
		ID:              1,
		AccountID:       1,
		ContractAddress: "0xtestContract",
		AccountAddress:  "0xtestAccount",
		Balance:         7,
	}

	err := tokenBalanceTable.Insert(context.Background(), []TokenBalance{tokenBalanceAccount})
	require.NoError(t, err)

	it := db.Backend().NewIter(&pebble.IterOptions{
		LowerBound: []byte{byte(TokenBalanceTableID)},
		UpperBound: []byte{byte(TokenBalanceTableID + 1)},
	})

	for it.First(); it.Valid(); it.Next() {
		rawData := it.Value()

		var tokenBalanceAccount1FromDB TokenBalance
		err = tokenBalanceTable.Serializer().Deserialize(rawData, &tokenBalanceAccount1FromDB)
		require.NoError(t, err)
		assert.Equal(t, tokenBalanceAccount, tokenBalanceAccount1FromDB)
	}

	_ = it.Close()

	tableUnsafeUpdater, ok := tokenBalanceTable.(TableUnsafeUpdater[TokenBalance])
	require.True(t, ok)

	err = tableUnsafeUpdater.UnsafeUpdate(
		context.Background(),
		[]TokenBalance{tokenBalanceAccountUpdated},
		[]TokenBalance{tokenBalanceAccount},
	)
	require.NoError(t, err)

	it = db.Backend().NewIter(&pebble.IterOptions{
		LowerBound: []byte{byte(TokenBalanceTableID)},
		UpperBound: []byte{byte(TokenBalanceTableID + 1)},
	})

	for it.First(); it.Valid(); it.Next() {
		rawData := it.Value()

		var tokenBalanceAccount1FromDB TokenBalance
		err = tokenBalanceTable.Serializer().Deserialize(rawData, &tokenBalanceAccount1FromDB)
		require.NoError(t, err)
		assert.Equal(t, tokenBalanceAccountUpdated, tokenBalanceAccount1FromDB)
	}

	_ = it.Close()
}
