import re

class ComplianceChecker:
    @staticmethod
    def check(text, config, doc_type):
        risks = []
        rules = config.get('document_types', {}).get(doc_type, {})
        required_fields = rules.get('required_fields', [])
        risk_flags = rules.get('risk_flags', [])

        for field in required_fields:
            pattern = re.compile(re.escape(field), re.IGNORECASE)
            if not pattern.search(text):
                risks.append(f'missing_{field}')

        for risk in risk_flags:
            if risk.startswith('missing_') and risk not in risks:
                continue
        return risks