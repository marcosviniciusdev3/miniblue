# Missing Azure Services & Features Analysis for miniblue

This document provides a comprehensive analysis of the services, features, and capabilities missing in `miniblue` that are critical for local Azure emulation, specifically focusing on enabling Infrastructure as Code (Terraform / Bicep) and common application deployment workflows without failure.

---

## 1. Executive Summary

`miniblue` emulates 28 Azure services with high performance (<1s startup, tiny footprint). However, while it provides basic resource group, networking (VNet/Subnet), storage, and VM stubs, modern Azure Infrastructure as Code (IaC) configurations—such as Terraform's `azurerm` provider or Bicep templates—frequently depend on lower-level infrastructure building blocks, governance resources, and data plane mechanisms that are currently missing or incomplete.

---

## 2. Categorized Breakdown of Missing Services & Features

### A. Compute & Core Infrastructure (Highest IaC Impact)
* **Network Interfaces (`Microsoft.Network/networkInterfaces`)**: [CRITICAL]
  * *Gap*: Essential bridge between VMs/VNets and IP configurations. Terraform `azurerm_linux_virtual_machine` and `azurerm_network_interface` almost always provision NICs first.
* **Managed Disks (`Microsoft.Compute/disks`)**: [CRITICAL]
  * *Gap*: Required for OS and data disks (`azurerm_managed_disk`). VMs currently assume inline storage configuration, causing Terraform disk resource provisioning to fail.
* **App Service / Web Apps (`Microsoft.Web/sites`)**: [HIGH]
  * *Gap*: App Service plans and Web Apps are standard for hosting web services. Currently, only basic Function app handlers exist.

### B. Governance & Organizational Scope
* **Management Groups (`Microsoft.Management/managementGroups`)**: [HIGH]
  * *Gap*: Subscriptions are often organized in management group hierarchies in Terraform enterprise modules.
* **Azure Policy (`Microsoft.Authorization/policyDefinitions`, `policyAssignments`)**: [HIGH]
  * *Gap*: Needed for testing compliance-as-code and Bicep/Terraform guardrails locally.
* **RBAC & Role Assignments (`Microsoft.Authorization/roleAssignments`)**: [MEDIUM]
  * *Gap*: Infrastructure modules frequently assign `Key Vault Secrets User`, `Storage Blob Data Contributor`, etc., to managed identities.

### C. Advanced Networking & Edge
* **Private Endpoints & Private DNS Zones (`Microsoft.Network/privateEndpoints`, `privateDnsZones`)**: [HIGH]
  * *Gap*: Standard secure cloud architecture uses private endpoints for PaaS services (Blob, Key Vault, SQL).
* **Network Interfaces & Route Tables (`Microsoft.Network/routeTables`)**: [MEDIUM]
  * *Gap*: Custom routing rules for VNet traffic.

### D. Cloud Native & Application Services
* **Azure Container Apps (`Microsoft.App/containerApps`)**: [HIGH]
  * *Gap*: Rapidly replacing ACI/AKS for modern microservices.
* **Event Hubs / API Management**: [MEDIUM]
  * *Gap*: Event stream ingestion and API gateways.

---

## 3. Control Plane vs. Data Plane Gap Analysis

| Service Domain | ARM Control Plane Gap | Data Plane Gap |
| :--- | :--- | :--- |
| **Compute / VM** | Missing `networkInterfaces` and `disks` resource endpoints. | Basic container/process execution supported, but lacks serial console / metrics stream. |
| **Key Vault** | Supports secrets (`/secrets`). Missing keys (`/keys`) and certificates (`/certificates`). | Missing cryptographic sign/verify operations and key rotation. |
| **Messaging (Service Bus)**| Basic queues/topics created. | Missing topic subscriptions, dead-lettering, session handlers, and AMQP protocol emulation. |
| **Database / Cosmos DB** | Account and database creation supported. | Missing SQL query parsing (currently returns dummy items or basic key-value retrieval). |
| **Storage** | Basic blob/queue/table CRUD supported. | Missing SAS token generation/validation, append blobs, and leasing locks. |

---

## 4. Recommended Priority Additions

To maximize compatibility with standard Terraform `azurerm` providers and Bicep deployments, the immediate priorities for `miniblue` are:

1. **`Microsoft.Network/networkInterfaces`**: Full CRUD operations for NIC resources and IP configurations.
2. **`Microsoft.Compute/disks`**: Full CRUD operations for Managed Disks.
3. **`Microsoft.Management/managementGroups`**: Management group structure.
4. **`Microsoft.Authorization/policyAssignments`**: Policy assignment stubs.
