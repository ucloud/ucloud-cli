# UDNS Subcommand Design

Date: 2026-07-14
Branch: feature/udns

## Overview

Add a fully-functional `udns` subcommand to ucloud-cli that covers all 10 UDNS API operations from `ucloud-sdk-go v0.22.90`. Zone operations are top-level commands under `udns`; record operations are grouped under `udns record`.

## Command Structure (Option A — flat zones, grouped records)

```
udns create            CreateUDNSZone
udns list              DescribeUDNSZone
udns modify            ModifyUDNSZone
udns associate-vpc     AssociateUDNSZoneVPC
udns disassociate-vpc  DisassociateUDNSZoneVPC
udns record list       DescribeUDNSRecord
udns record create     CreateUDNSRecord
udns record modify     ModifyUDNSRecord
udns record delete     DeleteUDNSRecord
```

`DescribeUDNSDomain` (list all RR records by zone name + VPC) is not exposed as a separate command; `udns record list` (DescribeUDNSRecord by zone ID) covers the primary use case.

## File Structure

```
products/udns/
  product.go           (existing — updated to mount all 6 top-level commands)
  create.go            NewCreateCommand      → CreateUDNSZone
  list.go              newListCommand        → DescribeUDNSZone
  modify.go            newModifyCommand      → ModifyUDNSZone
  associate_vpc.go     newAssociateVPCCommand    → AssociateUDNSZoneVPC
  disassociate_vpc.go  newDisassociateVPCCommand → DisassociateUDNSZoneVPC
  record.go            newRecordCommand      → root of `udns record` subgroup
  record_list.go       newRecordListCommand  → DescribeUDNSRecord
  record_create.go     newRecordCreateCommand → CreateUDNSRecord
  record_modify.go     newRecordModifyCommand → ModifyUDNSRecord
  record_delete.go     newRecordDeleteCommand → DeleteUDNSRecord
  rows.go              ZoneRow, RecordRow table types
```

`zone.go` (existing stub exporting `NewZoneCommand`) is deleted — `product.go` no longer calls it.

## Command Flags

### `udns create`
| Flag | Required | Description |
|---|---|---|
| `--zone-name` | yes | Domain name string |
| `--type` | yes | "private" or "public" (currently only private supported) |
| `--charge-type` | no | Year/Month/Dynamic, default Month |
| `--quantity` | no | Purchase duration, default 1 |
| `--recursion` | no | "enable" or "disable" |
| `--tag` | no | Business group |
| `--remark` | no | Remark |

Output: prints created `DNSZoneId`.

### `udns list`
| Flag | Required | Description |
|---|---|---|
| `--zone-id` | no | Filter by zone ID (repeatable) |
| `--offset` | no | Pagination offset, default 0 |
| `--limit` | no | Pagination limit, default 20 |

Output: table with columns ZoneID, Name, ChargeType, Recursion, VPCs, Tag, Remark, CreateTime, ExpireTime.

### `udns modify`
| Flag | Required | Description |
|---|---|---|
| `--zone-id` | yes | Zone resource ID |
| `--recursion` | no | "enable" or "disable" |
| `--remark` | no | Remark |

Output: success message. (SDK only exposes recursion and remark for modification.)

### `udns associate-vpc`
| Flag | Required | Description |
|---|---|---|
| `--zone-id` | yes | Zone resource ID |
| `--vpc-id` | yes | VPC resource ID |
| `--vpc-project-id` | yes | Project ID that owns the VPC |

Output: success message.

### `udns disassociate-vpc`
| Flag | Required | Description |
|---|---|---|
| `--zone-id` | yes | Zone resource ID |
| `--vpc-id` | yes | VPC resource ID |
| `--vpc-project-id` | yes | Project ID that owns the VPC |

Output: success message.

### `udns record list`
| Flag | Required | Description |
|---|---|---|
| `--zone-id` | yes | Zone resource ID |
| `--record-id` | no | Filter by record ID (repeatable) |
| `--query` | no | Fuzzy search string |
| `--offset` | no | Pagination offset, default 0 |
| `--limit` | no | Pagination limit, default 20 |

Output: table with columns RecordID, Name, Type, TTL, Values, ValueType, Remark.

### `udns record create`
| Flag | Required | Description |
|---|---|---|
| `--zone-id` | yes | Zone resource ID |
| `--name` | yes | Host record (subdomain) |
| `--type` | yes | Record type: A/AAAA/CNAME/MX/TXT/SRV/PTR |
| `--value` | yes | Value string: `IP\|weight\|enabled,...` |
| `--value-type` | yes | "Normal" or "Multivalue" |
| `--ttl` | no | TTL in seconds (5–600), default 5 |
| `--remark` | no | Remark |

Output: prints created `DNSRecordId`.

### `udns record modify`
| Flag | Required | Description |
|---|---|---|
| `--zone-id` | yes | Zone resource ID |
| `--record-id` | yes | Record resource ID |
| `--type` | no | Record type |
| `--value` | no | Value string |
| `--value-type` | no | "Normal" or "Multivalue" |
| `--ttl` | no | TTL in seconds |
| `--remark` | no | Remark |

Output: success message.

### `udns record delete`
| Flag | Required | Description |
|---|---|---|
| `--zone-id` | yes | Zone resource ID |
| `--record-id` | yes | Record resource ID (repeatable) |

Output: success message.

## Data Flow

Each command:
1. Calls `cli.NewServiceClient(ctx, udns.NewClient)` at construction time to get a typed SDK client.
2. Builds the SDK request struct and binds CLI flags onto its fields.
3. Binds `region` and `project-id` via `ctx.BindRegion` / `ctx.BindProjectID` (all UDNS requests embed `CommonBase`).
4. In `Run`: calls the SDK method; on error calls `ctx.HandleError(err)` and returns.
5. On success: `ctx.PrintList(rows)` for list commands; prints the returned resource ID for create commands; prints a success info line for modify/delete/associate commands.

## Table Row Types (`rows.go`)

```go
type ZoneRow struct {
    ZoneID     string
    Name       string
    ChargeType string
    Recursion  string
    VPCs       string
    Tag        string
    Remark     string
    CreateTime string
    ExpireTime string
}

type RecordRow struct {
    RecordID  string
    Name      string
    Type      string
    TTL       string
    Values    string
    ValueType string
    Remark    string
}
```

`VPCs` is a comma-joined list of VPC IDs from `ZoneInfo.VPCInfos`.
`Values` is a formatted string of the `ValueSet` entries.

## Out of Scope

- `DescribeUDNSDomain` — overlaps with `udns record list`; not exposed as a CLI command.
- Delete zone — no `DeleteUDNSZone` API exists in the SDK.
- Auto-renewal configuration — not in the SDK.
