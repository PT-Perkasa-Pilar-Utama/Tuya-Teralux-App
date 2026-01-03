# Backend Revision Plan

This document outlines the pending revisions required to fix the "Permission Deny (1106)" error by ensuring the backend correctly stores and uses Tuya Device IDs.

## 1. Database Schema Refactor
- **Target File**: `backend/domain/teralux/entities/device_entity.go`
- **Action**: Modify `Device` struct to include comprehensive Tuya fields.
- **Fields to Add**:
    - `TuyaID` (string, index)
    - `RemoteID` (string, index)
    - `Category` (string)
    - `RemoteCategory` (string)
    - `ProductName` (string)
    - `RemoteProductName` (string)
    - `LocalKey` (string)
    - `GatewayID` (string)
    - `IP` (string)
    - `Model` (string)
    - `Icon` (string)
- **Migration**: Ensure GORM `AutoMigrate` runs on startup to update the table schema.

## 2. Sync Logic Refactor (Auto-Upsert)
- **Target File**: `backend/domain/tuya/usecases/sync_device_status_usecase.go`
- **Action**: Update the `Execute` method to save fetched Tuya data into the database.
- **Logic**:
    1.  Fetch all devices from Tuya API.
    2.  Iterate through the list (including flattened Collections).
    3.  **Upsert**:
        - Check if device exists by `TuyaID` or `RemoteID`.
        - If exists: Update all fields (Name, Online, LocalKey, etc.).
        - If new: Create a new record with a new Teralux UUID.

## 3. Control/Fetch Logic Refactor
- **Target File**: `backend/domain/tuya/usecases/get_device_by_id_usecase.go` (and related controllers)
- **Action**: Update logic to map Teralux UUID to Tuya ID before calling external APIs.
- **Logic**:
    1.  Receive request with Teralux UUID.
    2.  Query Database for this UUID.
    3.  Extract the stored `TuyaID` (or `RemoteID`) from the record.
    4.  Use this valid ID to call the Tuya Open API.

## Summary
Completing these steps will ensure that the backend relies on its own robust database for device mapping, preventing the use of internal UUIDs for external API calls, thus resolving the permission error.
