package api

import "time"

func demoConsoleState() map[string]any {
	return map[string]any{
		"account": map[string]any{
			"id":     "acct-demo",
			"name":   "演示实验室",
			"status": "active",
		},
		"wallet": map[string]any{
			"balance":        2500,
			"frozen":         420,
			"totalRecharged": 5000,
		},
		"packages": []map[string]any{
			{
				"id":        "basic",
				"name":      "基础工作区",
				"server":    "标准 CPU",
				"cpu":       2,
				"memoryGb":  4,
				"diskGb":    10,
				"available": true,
				"price":     map[string]any{"computeHourly": 0.39, "storageGbMonth": 0.36},
			},
			{
				"id":        "pro",
				"name":      "专业工作区",
				"server":    "标准 CPU",
				"cpu":       8,
				"memoryGb":  16,
				"diskGb":    100,
				"available": true,
				"price":     map[string]any{"computeHourly": 3.09, "storageGbMonth": 0.36},
			},
		},
		"computePools":       demoComputePools(),
		"computeAllocations": demoComputeAllocations(),
		"storageVolumes": []map[string]any{
			{
				"id":        "vol-alpha",
				"name":      "Alpha 数据卷",
				"sizeGb":    100,
				"status":    "available",
				"accountId": "acct-demo",
			},
		},
		"storageAttachments": []map[string]any{
			{
				"id":        "att-alpha",
				"computeId": "cmp-alpha",
				"storageId": "vol-alpha",
				"mountPath": "/data",
				"status":    "attached",
			},
		},
		"workspaces": []map[string]any{
			{
				"id":                  "ws-alpha",
				"name":                "Alpha 实验室",
				"state":               "running",
				"status":              "running",
				"runtimeStatus":       "ready",
				"url":                 "https://workspace.example/ws-alpha",
				"packageId":           "basic",
				"computeId":           "cmp-alpha",
				"computeAllocationId": "cmp-alpha",
				"storageId":           "vol-alpha",
				"attachmentId":        "att-alpha",
				"access":              map[string]any{"tokenStatus": "active"},
				"billingAccount":      "acct-demo",
				"createdAt":           time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339),
			},
		},
		"manualTopups": []map[string]any{
			{"id": "topup-demo-1", "targetAccountId": "acct-demo", "amount": 5000, "reason": "初始额度"},
		},
		"walletTransactions": []map[string]any{
			{"id": "txn-demo-1", "type": "topup", "accountId": "acct-demo", "amount": 5000},
			{"id": "txn-demo-2", "type": "hold", "workspaceId": "ws-alpha", "amount": -420},
		},
		"billingLedger":     []map[string]any{},
		"resourceUsageLogs": []map[string]any{},
		"requestUsageLogs":  []map[string]any{},
		"notifications": []map[string]any{
			{"id": "note-demo-1", "type": "runtime", "severity": "info", "workspaceId": "ws-alpha"},
		},
	}
}

func demoComputePools() []map[string]any {
	return []map[string]any{
		{
			"id":        "pool-standard-cpu",
			"name":      "标准 CPU",
			"cpu":       8,
			"memoryGb":  32,
			"status":    "available",
			"region":    "local",
			"available": true,
		},
	}
}

func demoComputeAllocations() []map[string]any {
	return []map[string]any{
		{
			"id":        "cmp-alpha",
			"name":      "Alpha 计算",
			"poolId":    "pool-standard-cpu",
			"status":    "running",
			"cpu":       4,
			"memoryGb":  16,
			"accountId": "acct-demo",
		},
	}
}

func demoOperatorSummary() map[string]any {
	return map[string]any{
		"accounts":           map[string]any{"total": 1, "frozen": 420},
		"workspaces":         map[string]any{"total": 1, "running": 1},
		"computeAllocations": map[string]any{"total": 1},
		"runtimeOperations":  map[string]any{"failed": 0},
		"notifications":      map[string]any{"total": 1, "error": 0, "recent": []map[string]any{}},
	}
}

func demoManagementState() map[string]any {
	return map[string]any{
		"organizations": []map[string]any{
			{"id": "org-demo", "name": "演示组织", "billingAccountId": "acct-demo", "status": "active"},
		},
		"accounts": []map[string]any{
			{"id": "acct-demo", "balance": 2500, "frozen": 420, "totalRecharged": 5000, "status": "active"},
		},
		"users": []map[string]any{
			{"id": "user-demo-owner", "email": "owner@opl.local", "name": "OPL 所有者", "role": "lab_owner", "accountId": "acct-demo", "status": "active"},
			{"id": "user-demo-admin", "email": "admin@opl.local", "name": "OPL 管理员", "role": "admin", "accountId": "acct-operator", "status": "active"},
		},
	}
}
