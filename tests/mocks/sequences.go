package mocks

import "github.com/Tsarbomba69-com/mammoth.server/models"

// ╔══════════════════════════════════════ Basic Mock ══════════════════════════════════════╗

var mockSequences = []models.Sequence{
	{
		Name:          "user_id_seq",
		SchemaName:    "public",
		StartValue:    1,
		MinValue:      1,
		MaxValue:      9223372036854775807,
		Increment:     1,
		IsCyclic:      false,
		OwnedByTable:  "users",
		OwnedByColumn: "id",
	},
	{
		Name:          "order_id_seq",
		SchemaName:    "public",
		StartValue:    1000,
		MinValue:      1000,
		MaxValue:      999999999,
		Increment:     1,
		IsCyclic:      false,
		OwnedByTable:  "orders",
		OwnedByColumn: "order_id",
	},
}

// ╚══════════════════════════════════════════════════════════════════════════════════════════╝

// ╔══════════════════════════════════════ Case 1: Identical Sequences ══════════════════════════════════════╗

var IdenticalSourceSchema = []models.Schema{
	{
		Name: "public",
		Sequences: []models.Sequence{
			{
				Name:       "seq1",
				SchemaName: "public",
				StartValue: 10,
			},
		},
	},
}

var IdenticalTargetSchema = []models.Schema{
	{
		Name: "public",
		Sequences: []models.Sequence{
			{
				Name:       "seq1",
				SchemaName: "public",
				StartValue: 10,
			},
		},
	},
}

// ╚═════════════════════════════════════════════════════════════════════════════════════════════════════════╝
