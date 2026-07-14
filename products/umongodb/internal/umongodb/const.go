package umongodb

// productName is the single source of truth for the umongodb command name
// and its resource-id flag (--umongodb-id).
const productName = "umongodb"

// resourceIDFlag is the resource-id flag, named after the product per the
// onboarding contract.
const resourceIDFlag = productName + "-id" // "umongodb-id"
