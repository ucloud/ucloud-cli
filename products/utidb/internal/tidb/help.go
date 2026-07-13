package tidb

const (
	helpUTiDBRoot = `Manage UTiDB (TiDB) cluster instances on UCloud.

Common enums:
  ServerType (node type): tidb, tikv, pd, tiflash (case-insensitive; sent to API as Tidb/Tikv/Pd/Tiflash)
  NodeCount: must be greater than 3 per node type (minimum 4)
  ScaleType: SCALEOUT (expand), SCALEIN (shrink; scale-node/resize-disk only)
  ChargeType: Month, Year, Dynamic, Trial
  DTType: 10 (same AZ), 20 (cross AZ)
  DbVersion: e.g. v8.5.1, v8.5.6 (use list-specs to see available specs per region)`

	helpCreateLong = `Create a UTiDB instance.

Repeat --node-config for each node type. Example:
  --node-config 'ConfigId=tidb_2c_4g,DiskSize=100,NodeCount=4,ServerType=tidb'
  --node-config 'ConfigId=tikv_4c_16g,DiskSize=200,NodeCount=4,ServerType=tikv'
  --node-config 'ConfigId=pd_2c_4g,DiskSize=50,NodeCount=4,ServerType=pd'

Use 'utidb list-specs --node-types tidb,tikv,pd' to discover ConfigId values.
NodeCount must be greater than 3 for every node type (minimum 4).`

	helpScaleNodeLong = `Scale nodes of a UTiDB instance.

ScaleType:
  SCALEOUT  Expand nodes; NodeCount is the target total count after scaling.
  SCALEIN   Shrink nodes; NodeCount is the target total count after scaling.
            Requires --server-id of the node to remove (use tab completion or GetTiDBClusterService).

NodeCount must remain greater than 3 per node type (minimum 4) after the operation.
Example SCALEOUT: --scale-type SCALEOUT --node-config 'ConfigId=tikv_4c_16g,NodeCount=4,ServerType=tikv'
Example SCALEIN:  --scale-type SCALEIN --server-id <uuid> --node-config 'ConfigId=tikv_4c_16g,NodeCount=4,ServerType=tikv'`

	helpResizeDiskLong = `Resize disk of a UTiDB instance.

ScaleType: SCALEOUT or SCALEIN (disk expansion or shrink per node type).
Example: --scale-type SCALEOUT --node-config 'DiskSize=300,ServerType=tikv'`

	helpModifySpecLong = `Modify uhost specs of a UTiDB instance.

Example: --node-config 'ConfigId=tidb_4c_8g,ServerType=tidb'
Use 'utidb list-specs' to discover ConfigId values per node type.`

	helpListSpecsLong = `List available uhost specs for UTiDB node types.

Node types: tidb, tikv, pd, tiflash (comma-separated).
Example: --node-types tidb,tikv,pd`
)
