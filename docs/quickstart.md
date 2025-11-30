# GitGrove Quick Start

Get up and running with GitGrove in 5 minutes.

## Scenario

You have a microservices monorepo:
```
ecommerce/
├── api/
├── web/
└── services/
    ├── auth/
    ├── payments/
    └── inventory/
```

## Step 1: Initialize

```bash
cd ecommerce/
gitgrove init
```

## Step 2: Register Services

```bash
# Register each service
gitgrove register --name api --path api
gitgrove register --name web --path web
gitgrove register --name auth --path services/auth
gitgrove register --name payments --path services/payments
```

## Step 3: Create Hierarchy (Optional)

```bash
# Link services to show they're related
gitgrove link --child auth --parent api
gitgrove link --child payments --parent api
```

## Step 4: Start Working

### Switch to a Service
```bash
gitgrove switch auth main
```
Your terminal is now "inside" just the auth service!

### Make Changes
```bash
echo "feature" >> login.go
git grove add .
gitgrove commit "Add login retry logic"
```

### Navigate
```bash
gitgrove ls           # See child repos
gitgrove cd payments  # Jump to payments
gitgrove cd ..        # Go back up
gitgrove info         # See full tree
```

## Next Steps

- Read the [full README](../README.md)
- Explore [internals](internals.md)
- Try `gitgrove interactive` for menu-driven mode
