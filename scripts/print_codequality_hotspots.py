#!/usr/bin/env python3

import json
from pathlib import Path


def main() -> None:
    report = Path("reports/codequality/weekly-scorecard.json")
    data = json.loads(report.read_text(encoding="utf-8"))

    

    rows = data.get("hotspots", [])[:10]
    print("Top 10 Hotspots")
    print(
        "{:<6}  {:<50}  {:>7}  {:>7}  {:>7}".format(
            "Risk", "File", "Churn", "Cyclo", "Cogn"
        )
    )
    print(
        "{:<6}  {:<50}  {:>7}  {:>7}  {:>7}".format(
            "------", "-" * 50, "-----", "-----", "----"
        )
    )
    for row in rows:
        print(
            "{risk:>6}  {file:<50}  {churn:>7}  {cyclo:>7}  {cogn:>7}".format(
                risk=row["risk"],
                file=row["file"][:50],
                churn=row["churn_180d"],
                cyclo=row["cyclomatic_sum"],
                cogn=row["cognitive_sum"],
            )
        )


if __name__ == "__main__":
    main()
