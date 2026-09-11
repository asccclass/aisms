from __future__ import annotations

import importlib.util
import sqlite3
import sys
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("import_firewall_excel.py")
spec = importlib.util.spec_from_file_location("import_firewall_excel", MODULE_PATH)
import_firewall_excel = importlib.util.module_from_spec(spec)
assert spec.loader is not None
sys.modules["import_firewall_excel"] = import_firewall_excel
spec.loader.exec_module(import_firewall_excel)


class ImportFirewallExcelTest(unittest.TestCase):
    def test_insert_rows_replaces_existing_duplicate(self) -> None:
        conn = sqlite3.connect(":memory:")
        conn.execute(
            """
            CREATE TABLE firewall_requests (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                legacy_form_number TEXT NOT NULL DEFAULT '',
                system_name TEXT NOT NULL DEFAULT '',
                action TEXT NOT NULL DEFAULT '',
                purpose_type TEXT NOT NULL DEFAULT '',
                source_zone TEXT NOT NULL DEFAULT '',
                source_zone2 TEXT NOT NULL DEFAULT '',
                source_ip TEXT NOT NULL DEFAULT '',
                destination_zone TEXT NOT NULL DEFAULT '',
                destination_zone2 TEXT NOT NULL DEFAULT '',
                destination_ip TEXT NOT NULL DEFAULT '',
                protocol_type TEXT NOT NULL DEFAULT '',
                start_date TEXT NOT NULL DEFAULT '',
                end_date TEXT NOT NULL DEFAULT '',
                request_date TEXT NOT NULL DEFAULT '',
                rule_description TEXT NOT NULL DEFAULT '',
                firewall_zone TEXT NOT NULL DEFAULT '',
                firewall_id TEXT NOT NULL DEFAULT '',
                status TEXT NOT NULL DEFAULT 'active',
                creator TEXT NOT NULL DEFAULT '',
                remarks TEXT NOT NULL DEFAULT ''
            )
            """
        )
        row = import_firewall_excel.FirewallRow(
            legacy_form_number="I-20260911-001",
            system_name="舊系統名稱",
            action="開通",
            purpose_type="常態性服務",
            source_zone="campus",
            source_zone2="",
            source_ip="10.0.0.1/32",
            destination_zone="資料中心",
            destination_zone2="E",
            destination_ip="10.0.0.2/32",
            protocol_type="TCP 443",
            start_date="2026-09-11",
            end_date="2027-09-15",
            request_date="2026-09-11",
            rule_description="舊規則說明",
            firewall_zone="機房:DC",
            firewall_id="D-260911-1-1",
        )
        inserted, replaced, excel_duplicates = import_firewall_excel.insert_rows(conn, [row], "owner@example.com")
        self.assertEqual((inserted, replaced, excel_duplicates), (1, 0, 0))

        row.system_name = "Excel 系統名稱"
        row.rule_description = "Excel 規則說明"
        inserted, replaced, excel_duplicates = import_firewall_excel.insert_rows(conn, [row], "owner@example.com")
        self.assertEqual((inserted, replaced, excel_duplicates), (1, 1, 0))

        rows = conn.execute("SELECT system_name, rule_description FROM firewall_requests").fetchall()
        self.assertEqual(rows, [("Excel 系統名稱", "Excel 規則說明")])


if __name__ == "__main__":
    unittest.main()
