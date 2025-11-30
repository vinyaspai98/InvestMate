# InvestMate: Financial Tracking Application - Project Instructions

This document outlines the requirements and technical specifications for developing **InvestMate**, a financial tracking Single-Page Application (SPA) focused on the Indian financial market.

## 1. Project Overview & Design

* **Project Name:** InvestMate
* **Goal:** To build a clean, card-based dashboard application for tracking individual net worth, encompassing Indian stocks, mutual funds, FDs, insurance, and loans.
* **Design Reference:** The application's layout, structure, and component styling **must** strictly follow the provided wireframe design:
    [https://uxpilot.ai/p/rf2g9IfblZbgRC9hijl0?wireframeId=T7Ep1G7T084jZfKIEyTe&fullscreen=true](https://uxpilot.ai/p/rf2g9IfblZbgRC9hijl0?wireframeId=T7Ep1G7T084jZfKIEyTe&fullscreen=true)

---

## 2. Technical Stack & Infrastructure

| Component | Technology | Notes |
| :--- | :--- | :--- |
| **Frontend** | **Angular** (with TypeScript) | Angular material UI. |
| **Backend** | **GoLang** | For API development and business logic. |
| **Database** | **Firebase** | Used for Authentication and data persistence. |
| **Styling** | Custom CSS/SASS | Must implement a **Dark/Light Mode** theme toggle. |
| **Data** | Mocked Data | Initial frontend development must use **mocked data** for portfolio values, charts, and transaction tables. |

---

## 3. Application Pages & Features

### 3.1. Authentication Module (`/login`, `/signup`)

* **Layout:** Centered card form.
* **Fields:** Email, Password.
* **Functionality:** Implement a toggle to switch between **Login** and **Signup** modes.
* **Security:** Include a **Password Reset** option.
* **Integration:** Must integrate with **Firebase Authentication**.

### 3.2. Core Layout & Navigation

* **Navigation:** Implement a persistent **Sidebar Navigation** on the left.
* **Sidebar Links:** **Dashboard**, **Investments**, **Loans**, **Profile**, **About Us**.

### 3.3. Homepage (Dashboard) (`/dashboard`)

* **Header Card:** Prominently display **Total Assets (Net Worth)**, calculated as:
    $$\text{Net Worth} = (\text{Investments} + \text{FDs} + \text{Insurance Value}) - (\text{Total Loans})$$
* **Category Summary Cards:** Display small, responsive cards for each category: **Stocks**, **Mutual Funds**, **FDs**, **Insurance**, **Loans**.
* **Card Metrics:** Each card must show:
    * Total **Invested Amount** / Principal Amount (for Loans).
    * **Current Value** / Outstanding Balance (for Loans).
    * **Profit/Loss Percentage** ($\%\text{P/L}$).
* **Navigation:** Clicking any category card **must** navigate to its corresponding **Investment Detail Page**.

### 3.4. Investment Detail Pages (`/investments/:category`)

*(This structure is common for all five categories: Stocks, Mutual Funds, FDs, Loans, and Insurance.)*

1.  **Summary Panel:** Top section displaying aggregate metrics: Total Invested/Principal, Current Value/Outstanding, and $\%\text{P/L}$ for the selected category.
2.  **Visualization:** Include a prominent **Line Chart** for historical tracking.
    * **Filters:** Implement time-series filters: **1M**, **3M**, **1Y**, **All**.
3.  **Data Table:** A detailed **Transaction/History Table** listing all individual transactions (buys, sells, deposits, payments, etc.) for that category.
4.  **Floating Action Button (FAB):** A floating **“+ Add Investment”** button.
    * **Interaction:** Clicking the FAB opens a **Modal/Popup Form** for data entry.
    * **Form Fields:** Must include common fields (Amount, Date, Notes) and category-specific fields (e.g., Stock Ticker, Mutual Fund Name, Interest Rate, Tenor).

### 3.5. Profile Page (`/profile`)

* **Personal Details:** Display and allow editing of Name, Email, and Contact information.
* **Security:** Provide an option to **Change Password**.
* **Preferences:** Implement settings for:
    * Default **Currency** (INR).
    * **Dark/Light Mode Toggle**.
    * Notification preferences (placeholder).

### 3.6. About Us Page (`/about`)

* **Content:** Minimal layout featuring the Mission Statement, Tagline, and a Contact Link.

---

## 4. Deliverables & Process Focus

The primary deliverable for the frontend is the **Angular component structure** and **routing** that strictly adheres to the provided wireframe. Focus on **clean design**, **sidebar-based navigation**, and **user-friendly interactions** via modals for all data entry. The initial iteration should prioritize the UX/UI with mocked data.