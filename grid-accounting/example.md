### Initial Balances

**Alice's Balance Sheet (Maker)**

| **Assets (Debits)**         | **Liabilities (Credits)**   |
| ---                         | ---                         |
| D 1000 Alice-coins          |                             |
|                             | Equity: C 1000 Alice-coins  |
| **Total**: 1000 Alice-coins | **Total**: 1000 Alice-coins |

**Bob's Balance Sheet (Taker)**

| **Assets (Debits)**      | **Liabilities (Credits)** |
| ---                      | ---                       |
| D 500 Bob-coins            |                           |
|                          | Equity: 500 Bob-coins     |
| **Total**: 500 Bob-coins | **Total**: 500 Bob-coins  |

### After Match (Promise Accepted)

**Alice's Balance Sheet (Maker)**

| **Assets (Debits)**   | **Liabilities (Credits)** |
| ---                   | ---                       |
| D 1000 Alice-coins    | C 1 Promise X (Liability) |
|                       |                           |
|                       | Equity: 1000 Alice-coins  |
|                       | Equity: 1 Promise X       |
| **Total**: 1001 units | **Total**: 1001 units     |

**Bob's Balance Sheet (Taker)**

| **Assets (Debits)**  | **Liabilities (Credits)** |
| ---                  | ---                       |
| 500 Bob-coins        | 100 Bob-coins (Liability) |
| 1 Promise X (Asset)  |                           |
|                      | Equity: 400 Bob-coins     |
| **Total**: 501 units | **Total**: 501 units      |

### After Fulfillment Declaration (Promise Fulfilled)

**Alice's Balance Sheet (Maker)**

| **Assets (Debits)**   | **Liabilities (Credits)** |
| ---                   | ---                       |
| 1000 Alice-coins      |                           |
| 100 Bob-coins (Asset) |                           |
|                       | Equity: 1000 Alice-coins  |
| **Total**: 1100 units | **Total**: 1000 units     |

**Bob's Balance Sheet (Taker)**

| **Assets (Debits)**  | **Liabilities (Credits)** |
| ---                  | ---                       |
| 500 Bob-coins        |                           |
| 1 Promise X (Asset)  | Equity: 400 Bob-coins     |
|                      | Credits: 100 Bob-coins    |
| **Total**: 501 units | **Total**: 501 units      |

