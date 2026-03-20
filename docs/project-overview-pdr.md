# Project Overview & PDR
## Health Data Platform

### 1. Introduction
The Health Data Platform (HDP) is a high-performance, secure, and extensible system designed for the ingestion and management of health data from generic smartwatch devices (utilizing the `IW` protocol). It provides a seamless user experience via a modern web interface for device registration and real-time data monitoring.

### 2. Product Development Requirements (PDR)
- **Scalability**: Capable of handling hundreds of persistent TCP connections concurrently and distributing data efficiently.
- **Security & Privacy**: User accounts are secured via Google OAuth 2.0. HMAC-signed sessions ensure cookie integrity. PII/PHI data is tracked via user-to-device ownership.
- **Extensibility**: The system is designed to handle multiple protocol variants, with a robust scanner that handles noise and delimiters.
- **Reliability**: A dual-server architecture separating the Echo HTTP API and the TCP Ingestion Server ensures that high-volume data ingestion doesn't impact user dashboard performance.

### 3. Target Audience
- **Healthcare Providers**: Monitoring patient health via registered smartwatch devices.
- **Data Analysts and Researchers**: Analyzing raw and normalized health data.
- **Patient Portals**: Patients managing their own devices and viewing health trackers.

### 4. Key Capabilities
- **Smartwatch Data Ingestion**: Robust TCP server (port 9090) for IW protocol ingestion with IMEI authentication.
- **User Presence & Authentication**: Google OAuth 2.0 login and persistent, signed sessions.
- **Interactive Web Dashboard**: Dynamic UI for device registration, IMEI validation, and real-time status monitoring.
- **Flexible Data Persistence**: Storing 13+ packet types (GPS, Heart rate, BP, Temperature, Sleep, etc.) into PostgreSQL with JSONB support.
