"""
End-to-end tests for CLI functionality.

These tests verify the complete CLI workflow from command parsing
to output generation.
"""

import pytest
from unittest.mock import patch, Mock
from io import StringIO
import sys
from kite import cli
from kite.utils.data_models import CaseData
from datetime import datetime


class TestCLISearchCommand:
    """E2E tests for search command."""

    @pytest.fixture
    def mock_cases(self):
        """Mock case data for testing."""
        return [
            CaseData(
                case_name="Test Case 1",
                case_id="TEST-001",
                court="Test Court",
                date=datetime(2023, 1, 15),
                url="https://example.com/case1",
                jurisdiction="US",
                citations=["123 Test 456"],
            ),
            CaseData(
                case_name="Test Case 2",
                case_id="TEST-002",
                court="Test Court",
                date=datetime(2023, 2, 20),
                url="https://example.com/case2",
                jurisdiction="US",
                citations=["789 Test 012"],
            ),
        ]

    def test_search_command_text_output(self, mock_cases, capsys):
        """Test search command with text output."""
        with patch("kite.cli.CourtListenerScraper") as MockScraper:
            mock_instance = MockScraper.return_value.__enter__.return_value
            mock_instance.search_cases.return_value = mock_cases

            # Simulate CLI arguments
            sys.argv = [
                "kite",
                "search",
                "courtlistener",
                "privacy",
                "--limit",
                "10",
            ]

            with patch("sys.argv", sys.argv):
                parser = cli.create_parser()
                args = parser.parse_args(sys.argv[1:])
                args.verbose = 0
                args.func = cli.search_command

                # Execute command
                args.func(args)

                # Check output
                captured = capsys.readouterr()
                assert "Test Case 1" in captured.out
                assert "Test Case 2" in captured.out

    def test_search_command_json_output(self, mock_cases, capsys):
        """Test search command with JSON output."""
        with patch("kite.cli.CourtListenerScraper") as MockScraper:
            mock_instance = MockScraper.return_value.__enter__.return_value
            mock_instance.search_cases.return_value = mock_cases

            sys.argv = [
                "kite",
                "search",
                "courtlistener",
                "privacy",
                "--format",
                "json",
                "--limit",
                "5",
            ]

            with patch("sys.argv", sys.argv):
                parser = cli.create_parser()
                args = parser.parse_args(sys.argv[1:])
                args.verbose = 0
                args.func = cli.search_command

                args.func(args)

                captured = capsys.readouterr()
                # Should contain JSON
                assert '"case_name"' in captured.out or "case_name" in captured.out

    def test_search_command_with_date_filters(self, mock_cases):
        """Test search command with date filters."""
        with patch("kite.cli.CourtListenerScraper") as MockScraper:
            mock_instance = MockScraper.return_value.__enter__.return_value
            mock_instance.search_cases.return_value = mock_cases

            sys.argv = [
                "kite",
                "search",
                "courtlistener",
                "privacy",
                "--start-date",
                "2023-01-01",
                "--end-date",
                "2023-12-31",
                "--limit",
                "10",
            ]

            with patch("sys.argv", sys.argv):
                parser = cli.create_parser()
                args = parser.parse_args(sys.argv[1:])
                args.verbose = 0
                args.func = cli.search_command

                # Should not raise
                args.func(args)

                # Verify search_cases was called with correct parameters
                mock_instance.search_cases.assert_called_once()
                call_kwargs = mock_instance.search_cases.call_args[1]
                assert call_kwargs["start_date"] == "2023-01-01"
                assert call_kwargs["end_date"] == "2023-12-31"

    def test_search_command_unknown_scraper(self, capsys):
        """Test search command with unknown scraper."""
        sys.argv = ["kite", "search", "unknown_scraper", "query"]

        with patch("sys.argv", sys.argv):
            parser = cli.create_parser()
            args = parser.parse_args(sys.argv[1:])
            args.verbose = 0
            args.func = cli.search_command

            with pytest.raises(SystemExit):
                args.func(args)

            captured = capsys.readouterr()
            assert "Unknown scraper" in captured.err

    def test_search_command_no_results(self, capsys):
        """Test search command when no results found."""
        with patch("kite.cli.CourtListenerScraper") as MockScraper:
            mock_instance = MockScraper.return_value.__enter__.return_value
            mock_instance.search_cases.return_value = []

            sys.argv = ["kite", "search", "courtlistener", "nonexistent"]

            with patch("sys.argv", sys.argv):
                parser = cli.create_parser()
                args = parser.parse_args(sys.argv[1:])
                args.verbose = 0
                args.func = cli.search_command

                args.func(args)

                captured = capsys.readouterr()
                assert "No cases found" in captured.out


class TestCLIGetCaseCommand:
    """E2E tests for get-case command."""

    @pytest.fixture
    def mock_case(self):
        """Mock single case data."""
        return CaseData(
            case_name="Test Case",
            case_id="TEST-001",
            court="Test Court",
            date=datetime(2023, 1, 15),
            url="https://example.com/case",
            jurisdiction="US",
            judges=["Judge Smith"],
            citations=["123 Test 456"],
            summary="This is a test case.",
        )

    def test_get_case_command_success(self, mock_case, capsys):
        """Test get-case command successfully retrieves a case."""
        with patch("kite.cli.CourtListenerScraper") as MockScraper:
            mock_instance = MockScraper.return_value.__enter__.return_value
            mock_instance.get_case_by_id.return_value = mock_case

            sys.argv = ["kite", "get-case", "courtlistener", "TEST-001"]

            with patch("sys.argv", sys.argv):
                parser = cli.create_parser()
                args = parser.parse_args(sys.argv[1:])
                args.verbose = 0
                args.func = cli.get_case_command

                args.func(args)

                captured = capsys.readouterr()
                assert "Test Case" in captured.out
                assert "TEST-001" in captured.out

    def test_get_case_command_not_found(self, capsys):
        """Test get-case command when case not found."""
        with patch("kite.cli.CourtListenerScraper") as MockScraper:
            mock_instance = MockScraper.return_value.__enter__.return_value
            mock_instance.get_case_by_id.return_value = None

            sys.argv = ["kite", "get-case", "courtlistener", "NONEXISTENT"]

            with patch("sys.argv", sys.argv):
                parser = cli.create_parser()
                args = parser.parse_args(sys.argv[1:])
                args.verbose = 0
                args.func = cli.get_case_command

                args.func(args)

                captured = capsys.readouterr()
                assert "not found" in captured.out


class TestCLIListScrapersCommand:
    """E2E tests for list-scrapers command."""

    def test_list_scrapers_command(self, capsys):
        """Test list-scrapers command outputs available scrapers."""
        sys.argv = ["kite", "list-scrapers"]

        with patch("sys.argv", sys.argv):
            parser = cli.create_parser()
            args = parser.parse_args(sys.argv[1:])
            args.func = cli.list_scrapers_command

            args.func(args)

            captured = capsys.readouterr()
            # Should list various scrapers
            assert "courtlistener" in captured.out.lower()
            assert "canlii" in captured.out.lower()
            assert "bailii" in captured.out.lower()


class TestCLIArgumentParsing:
    """Test CLI argument parsing."""

    def test_parse_search_arguments(self):
        """Test parsing search command arguments."""
        sys.argv = [
            "kite",
            "search",
            "courtlistener",
            "privacy",
            "--limit",
            "10",
            "--format",
            "json",
        ]

        with patch("sys.argv", sys.argv):
            parser = cli.create_parser()
            args = parser.parse_args(sys.argv[1:])

            assert args.command == "search"
            assert args.scraper == "courtlistener"
            assert args.query == "privacy"
            assert args.limit == 10
            assert args.format == "json"

    def test_parse_get_case_arguments(self):
        """Test parsing get-case command arguments."""
        sys.argv = [
            "kite",
            "get-case",
            "courtlistener",
            "12345",
            "--format",
            "text",
        ]

        with patch("sys.argv", sys.argv):
            parser = cli.create_parser()
            args = parser.parse_args(sys.argv[1:])

            assert args.command == "get-case"
            assert args.scraper == "courtlistener"
            assert args.case_id == "12345"
            assert args.format == "text"

    def test_parse_version(self):
        """Test version argument."""
        sys.argv = ["kite", "--version"]

        with patch("sys.argv", sys.argv):
            parser = cli.create_parser()

            with pytest.raises(SystemExit) as exc_info:
                parser.parse_args(sys.argv[1:])

            # Version should exit with code 0
            assert exc_info.value.code == 0

