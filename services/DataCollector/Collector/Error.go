package Collector

import "github.com/grdvsng/stock-analysis/Exceptions"

var ContractNotFoundError = Exceptions.CreateError("Contract for %s not found")

var StorageError = Exceptions.CreateError("Storage error for %s: %s")

var ReprocessError = Exceptions.CreateError("Reprocess of %s failed: %s")

var SourceError = Exceptions.CreateError("Source error for %s: %s")
