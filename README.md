# 🚀 DevOps Autopilot — Crossplane Packages

> Crossplane compositions & functions packaged as OCI artifacts to power the DevOps Autopilot platform.

---

## 🧠 Overview

This repository contains **platform-level abstractions built with Crossplane**.

It defines **custom infrastructure APIs (XRDs)**, their **implementations (Compositions)**, and optional **Functions**, all packaged as **OCI artifacts** and consumed by the main platform:

👉 `devops-autopilot` (EKS + ArgoCD + GitOps)

---

## 🎯 Objective

The goal of this repository is to **abstract infrastructure complexity behind clean, reusable platform APIs**.

Instead of exposing raw cloud resources (RDS, S3, IAM, etc.), this repo enables:

- Standardized infrastructure provisioning
- Self-service capabilities for teams
- Consistent, policy-driven deployments
- Versioned and reproducible platform components

---

## 🧩 Architecture Position

This repository represents the **platform API layer** in the DevOps Autopilot ecosystem.

High-level flow:

- Users interact with platform APIs (claims)
- Crossplane interprets those APIs
- This repository defines how those APIs map to infrastructure
- Cloud resources are provisioned automatically

This creates a clear separation between:

- Infrastructure provisioning (handled elsewhere)
- Platform abstraction (handled here)

---

## 📦 What This Repository Contains

### 1. XRDs (CompositeResourceDefinitions)
Defines the **APIs exposed to users**

These represent the contract between the platform and its consumers.

---

### 2. Compositions
Defines **how each API is implemented**

They map high-level platform APIs to:
- Cloud resources
- Kubernetes resources
- External systems

---

### 3. Functions (Optional)
Encapsulates reusable logic such as:

- Defaulting values
- Enforcing conventions
- Dynamic rendering
- Policy controls

---

### 4. Packages (OCI Artifacts)
All resources are bundled into:

- Crossplane packages (`xpkg`)
- Published to an OCI registry (e.g. GHCR)

These packages are then consumed by the main platform via GitOps.

---

## ⚙️ Development Workflow

The typical workflow follows these stages:

1. Define a platform API (XRD)
2. Implement its behavior via a Composition
3. Optionally add Functions for logic or defaults
4. Validate locally in a Kubernetes environment
5. Package everything into an OCI artifact
6. Publish via CI/CD
7. Consume from the main platform repository

---

## 🔁 CI/CD Pipeline

The repository is designed to be fully automated.

The pipeline will:

- Detect changes in APIs, compositions, or functions
- Build Crossplane packages
- Publish them to an OCI registry
- Optionally trigger updates in the platform repository

---

## 🧠 Design Principles

- **API-first platform design**
- **Separation of concerns (infra vs platform abstraction)**
- **No direct exposure of cloud resources**
- **GitOps compatibility**
- **Versioned and reproducible artifacts**
- **Extensibility for future platform services**

---

## 🚀 Roadmap

- Core infrastructure abstractions (databases, storage, identity)
- Function library for defaults and validation
- Multi-environment support
- Automated versioning and release management
- Deeper integration with DevOps Autopilot platform

---

## 🤝 Contribution

1. Create a feature branch  
2. Add or modify XRDs / Compositions / Functions  
3. Validate changes locally  
4. Open a pull request  
5. CI/CD handles packaging and publishing  

---

## ⚠️ Important Notes

- Do not commit generated artifacts (e.g. packages)
- Do not store secrets or credentials
- All outputs must be reproducible via CI/CD
- Treat this repository as a **source of truth for platform APIs**

---

## 🧠 Final Perspective

This repository is where:

> Infrastructure is transformed into a platform.

It defines how infrastructure is consumed — not how it is created.

---
