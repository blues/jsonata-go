package main

import "testing"

func TestReplaceQuotesInPaths(t *testing.T) {

	inputs := []string{
		`[Address, Other."Alternative.Address"].City`,
		`Account.(  $AccName := function() { $."Account Name" };  Order[OrderID = "order104"].Product{    "Account": $AccName(),    "SKU-" & $string(ProductID): $."Product Name"  } )`,
		`Account.Order.Product."Product Name".$uppercase().$substringBefore(" ")`,
		`"foo".**.fud`,
		`foo.**."fud"`,
		`"foo".**."fud"`,
		`Account.Order.Product[$."Product Name" ~> /hat/i].ProductID`,
		`$sort(Account.Order.Product."Product Name")`,
		`Account.Order.Product ~> $map(λ($prod, $index) { $index+1 & ": " & $prod."Product Name" })`,
		`Account.Order.Product ~> $map(λ($prod, $index, $arr) { $index+1 & "/" & $count($arr) & ": " & $prod."Product Name" })`,
		`Account.Order{OrderID: Product."Product Name"}`,
		`Account.Order.{OrderID: Product."Product Name"}`,
		`Account.Order.Product{$."Product Name": Price, $."Product Name": Price}`,
		`Account.Order{  OrderID: {    "TotalPrice":$sum(Product.(Price * Quantity)),    "Items": Product."Product Name"  }}`,
		`{  "Order": Account.Order.{      "ID": OrderID,      "Product": Product.{          "Name": $."Product Name",          "SKU": ProductID,          "Details": {            "Weight": Description.Weight,            "Dimensions": Description.(Width & " x " & Height & " x " & Depth)          }        },      "Total Price": $sum(Product.(Price * Quantity))    }}`,
		`Account.Order.Product[$contains($."Product Name", /hat/)].ProductID`,
		`Account.Order.Product[$contains($."Product Name", /hat/i)].ProductID`,
		`Account.Order.Product.$replace($."Product Name", /hat/i, function($match) { "foo" })`,
		`Account.Order.Product.$replace($."Product Name", /(h)(at)/i, function($match) { $uppercase($match.match) })`,
		`$.'7a'`,
		`$.'7'`,
		`$lowercase($."NI.Number")`,
		`$lowercase("COMPENSATION IS : " & Employment."Executive.Compensation")`,
		`Account[$$.Account."Account Name" = "Firefly"].*[OrderID="order104"].Product.Price`,
	}

	outputs := []string{
		"[Address, Other.`Alternative.Address`].City",
		"Account.(  $AccName := function() { $.`Account Name` };  Order[OrderID = \"order104\"].Product{    \"Account\": $AccName(),    \"SKU-\" & $string(ProductID): $.`Product Name`  } )",
		"Account.Order.Product.`Product Name`.$uppercase().$substringBefore(\" \")",
		"`foo`.**.fud",
		"foo.**.`fud`",
		"`foo`.**.`fud`",
		"Account.Order.Product[$.`Product Name` ~> /hat/i].ProductID",
		"$sort(Account.Order.Product.`Product Name`)",
		"Account.Order.Product ~> $map(λ($prod, $index) { $index+1 & \": \" & $prod.`Product Name` })",
		"Account.Order.Product ~> $map(λ($prod, $index, $arr) { $index+1 & \"/\" & $count($arr) & \": \" & $prod.`Product Name` })",
		"Account.Order{OrderID: Product.`Product Name`}",
		"Account.Order.{OrderID: Product.`Product Name`}",
		"Account.Order.Product{$.`Product Name`: Price, $.`Product Name`: Price}",
		"Account.Order{  OrderID: {    \"TotalPrice\":$sum(Product.(Price * Quantity)),    \"Items\": Product.`Product Name`  }}",
		"{  \"Order\": Account.Order.{      \"ID\": OrderID,      \"Product\": Product.{          \"Name\": $.`Product Name`,          \"SKU\": ProductID,          \"Details\": {            \"Weight\": Description.Weight,            \"Dimensions\": Description.(Width & \" x \" & Height & \" x \" & Depth)          }        },      \"Total Price\": $sum(Product.(Price * Quantity))    }}",
		"Account.Order.Product[$contains($.`Product Name`, /hat/)].ProductID",
		"Account.Order.Product[$contains($.`Product Name`, /hat/i)].ProductID",
		"Account.Order.Product.$replace($.`Product Name`, /hat/i, function($match) { \"foo\" })",
		"Account.Order.Product.$replace($.`Product Name`, /(h)(at)/i, function($match) { $uppercase($match.match) })",
		"$.`7a`",
		"$.`7`",
		"$lowercase($.`NI.Number`)",
		"$lowercase(\"COMPENSATION IS : \" & Employment.`Executive.Compensation`)",
		"Account[$$.Account.`Account Name` = \"Firefly\"].*[OrderID=\"order104\"].Product.Price",
	}

	for i := range inputs {

		got, ok := replaceQuotesInPaths(inputs[i])
		if got != outputs[i] {
			t.Errorf("\n     Input: %s\nExp. Output: %s\nAct. Output: %s", inputs[i], outputs[i], got)
		}
		if !ok {
			t.Errorf("%s: Expected true, got %t", inputs[i], ok)
		}
	}
}

func TestReplaceQuotesInPathsNoOp(t *testing.T) {

	inputs := []string{
		`42 ~> "hello"`,
		`"john@example.com" ~> $substringAfter("@") ~> $substringBefore(".")`,
		`$ ~> |Account.Order.Product|{"Total":Price*Quantity},["Description", "SKU"]|`,
		`$ ~> |(Account.Order.Product)[0]|{"Description":"blah"}|`,
	}

	for i := range inputs {

		got, ok := replaceQuotesInPaths(inputs[i])
		if got != inputs[i] {
			t.Errorf("\n     Input: %s\nExp. Output: %s\nAct. Output: %s", inputs[i], inputs[i], got)
		}
		if ok {
			t.Errorf("%s: Expected false, got %t", inputs[i], ok)
		}
	}
}
