from enum import Enum


class SupplierType(Enum):
    IBK = "ibk"
    YAHOO = "Yahoo"


ALLOWED_SUPPLIER_TYPES = {SupplierType.IBK.value, SupplierType.YAHOO.value}
