package suites

import (
	"fmt"

	"github.com/go-bond/bond"
	"github.com/go-bond/bond/_benchmarks/bench"
	"github.com/vmihailenco/msgpack/v5"
)

func init() {
	bench.RegisterBenchmarkSuite(
		bench.NewBenchmarkSuite("BenchmarkTableQueryWithTableSerializerSuite", "skip-table-query-with-table-serializer",
			BenchmarkTableQueryWithTableSerializerSuite),
	)
}

func BenchmarkTableQueryWithTableSerializerSuite(bs *bench.BenchmarkSuite) []bench.BenchmarkResult {
	msgpack.GetEncoder().SetCustomStructTag("json")
	msgpack.GetDecoder().SetCustomStructTag("json")

	var serializers = []struct {
		Name       string
		Serializer bond.Serializer[any]
	}{
		//{"JSONSerializer", &bond.JsonSerializer{}},
		{"MsgPackSerializer", &bond.MsgPackSerializer{}},
	}

	var results []bench.BenchmarkResult
	for _, serializer := range serializers {
		db := setupDatabase(serializer.Serializer)

		const (
			TokenBalanceTableID = bond.TableID(1)
		)

		tokenBalanceTable := bond.NewTable[*TokenBalance](db, TokenBalanceTableID, func(builder bond.KeyBuilder, tb *TokenBalance) []byte {
			return builder.AddUint64Field(tb.ID).Bytes()
		}, &TokenBalanceSerializer{})

		const (
			_                                                 = bond.PrimaryIndexID
			TokenBalanceAccountAddressIndexID                 = iota
			TokenBalanceAccountAddressOrderBalanceDESCIndexID = iota
			TokenBalanceAccountAndContractAddressIndexID      = iota
		)

		var (
			TokenBalanceAccountAddressIndex = bond.NewIndex[*TokenBalance](
				TokenBalanceAccountAddressIndexID,
				func(builder bond.KeyBuilder, tb *TokenBalance) []byte {
					return builder.AddStringField(tb.AccountAddress).Bytes()
				},
				bond.IndexOrderDefault[*TokenBalance],
			)
			TokenBalanceAccountAddressOrderBalanceDESCIndex = bond.NewIndex[*TokenBalance](
				TokenBalanceAccountAddressOrderBalanceDESCIndexID,
				func(builder bond.KeyBuilder, tb *TokenBalance) []byte {
					return builder.AddStringField(tb.AccountAddress).Bytes()
				},
				func(o bond.IndexOrder, tb *TokenBalance) bond.IndexOrder {
					return o.OrderUint64(tb.Balance, bond.IndexOrderTypeDESC)
				},
			)
			TokenBalanceAccountAndContractAddressIndex = bond.NewIndex[*TokenBalance](
				TokenBalanceAccountAndContractAddressIndexID,
				func(builder bond.KeyBuilder, tb *TokenBalance) []byte {
					return builder.
						AddStringField(tb.AccountAddress).
						AddStringField(tb.ContractAddress).
						Bytes()
				},
				bond.IndexOrderDefault[*TokenBalance],
			)
		)

		err := tokenBalanceTable.AddIndex([]*bond.Index[*TokenBalance]{
			TokenBalanceAccountAddressIndex,
			TokenBalanceAccountAddressOrderBalanceDESCIndex,
			TokenBalanceAccountAndContractAddressIndex,
		})
		if err != nil {
			panic(err)
		}

		var tokenBalances []*TokenBalance
		for i := 0; i < 20000000; i++ {
			tokenBalances = append(tokenBalances, &TokenBalance{
				ID:              uint64(i + 1),
				AccountID:       0,
				ContractAddress: "0xtestContract" + fmt.Sprintf("%d", i),
				AccountAddress:  "0xtestAccount0",
				Balance:         uint64((i % 100) * 10),
			})
		}

		err = tokenBalanceTable.Insert(tokenBalances[0:10000000])
		if err != nil {
			panic(err)
		}

		err = tokenBalanceTable.Insert(tokenBalances[10000000:20000000])
		if err != nil {
			panic(err)
		}

		var queryInputs = []struct {
			index     *bond.Index[*TokenBalance]
			indexName string
			selector  *TokenBalance
			filter    func(tb *TokenBalance) bool
			order     func(tb *TokenBalance, tb2 *TokenBalance) bool
			offset    int
			limit     int
		}{
			// Default Index
			{index: nil, indexName: "Default", selector: nil, offset: 0, limit: 0},
			{index: nil, indexName: "Default", selector: nil, offset: 0, limit: 500},
			{index: nil, indexName: "Default", selector: nil, offset: 0, limit: 1000},
			{index: nil, indexName: "Default", selector: nil, offset: 0, limit: 5000},
			{index: nil, indexName: "Default", selector: nil, offset: 0, limit: 10000},
			{index: nil, indexName: "Default", selector: nil, offset: 0, limit: 100000},
			{index: nil, indexName: "Default", selector: nil, offset: 0, limit: 0},
			{index: nil, indexName: "Default", selector: nil, offset: 500, limit: 1000},
			{index: nil, indexName: "Default", selector: nil, offset: 1000, limit: 1000},
			{index: nil, indexName: "Default", selector: nil, offset: 5000, limit: 1000},
			{index: nil, indexName: "Default", selector: nil, offset: 10000, limit: 1000},
			{index: nil, indexName: "Default", selector: nil, offset: 100000, limit: 1000},
			{index: nil, indexName: "Default", selector: nil, offset: 1000000, limit: 1000},
			{index: nil, indexName: "Default", selector: nil, offset: 10000000, limit: 1000},
			// AccountAddress Index
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 0},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 500},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 1000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 5000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 10000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 100000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 0},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 500, limit: 1000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 1000, limit: 1000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 5000, limit: 1000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 10000, limit: 1000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 100000, limit: 1000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 1000000, limit: 1000},
			{index: TokenBalanceAccountAddressIndex, indexName: "AccountAddress", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 10000000, limit: 1000},
			// AccountAddressOrderBalanceDESC
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 0},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 500},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 1000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 5000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 10000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 100000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 0, limit: 0},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 500, limit: 1000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 1000, limit: 1000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 5000, limit: 1000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 10000, limit: 1000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 100000, limit: 1000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 1000000, limit: 1000},
			{index: TokenBalanceAccountAddressOrderBalanceDESCIndex, indexName: "AccountAddressOrderBalanceDESC", selector: &TokenBalance{AccountAddress: "0xtestAccount0"}, offset: 10000000, limit: 1000},
		}

		for _, v := range queryInputs {
			results = append(results,
				bs.Benchmark(bench.Benchmark{
					Name: fmt.Sprintf("%s/%s/Query_Index_%s_Offset_%d_Limit_%d",
						bs.Name, serializer.Name, v.indexName, v.offset, v.limit),
					Inputs:        v,
					BenchmarkFunc: QueryWithOpts(tokenBalanceTable, v.index, v.selector, v.offset, v.limit),
				}),
			)
		}

		tearDownDatabase(db)
	}

	return results
}