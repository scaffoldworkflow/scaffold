from typing import List, Dict, Union
import scaffold.user, scaffold.workflow
from config import *
import uuid

def user_setup() -> str:
    test_id = str(uuid.uuid4())

    u = scaffold.user.User()
    u.loadf(USER_FIXTURE_PATH)
    u.username = test_id

    scaffold.user.create(u, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    return test_id

def user_teardown(test_id: str) -> None:
    scaffold.user.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)

def workflow_setup() -> str:
    test_id = user_setup()

    w = scaffold.workflow.Workflow()
    w.loadf(WORKFLOW_FIXTURE_PATH)
    w.name = test_id

    scaffold.workflow.create(w, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    return test_id

def workflow_teardown(test_id: str) -> None:
    scaffold.workflow.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    user_teardown(test_id)

# def setup_users():
#     for username in ["bar", "foo", "read-only", "no-group"]:
#         u = scaffold.user.User()
#         u.loadf(f'../fixtures/{username}.json')
#         status = scaffold.user.create(u, SCAFFOLD_BASE, SCAFFOLD_PRIMARY_KEY, fail_on_error=False)
#         assert status < 400

def get_letter_locations(contents: str, padding: str = " ") -> List[int]:
    """Gets the locations of starting letters of words in a string

    Args:
        contents (str): String to get starting letter locations from
        padding (str, optional): Padding between columns. Defaults to " ".

    Returns:
        List[int]: Indices in the string of starting letters of words
    """

    lines = contents.split("\n")

    # Remove empty lines
    lines = [line for line in lines if len(line) > 0][:1]

    max_length = 0
    letter_locations = []
    # Get all the locations of the first letter of a word
    for line in lines:
        line_letter_locations = []
        chars = list(line)
        first_flag = True
        for idx, char in enumerate(chars):
            # Is it checking for the beginning of a work and not a space
            if first_flag and char != padding:
                line_letter_locations.append(idx)
                first_flag = False
                continue
            # If it's a space then we want to look for the beginning of a word
            if char == padding:
                first_flag = True

        letter_locations.append(line_letter_locations)
        if len(line) > max_length:
            max_length = len(line)
    # Now that we have the letter locations we want to know which ones they all have in
    # common. Start by loading in the first line
    global_letter_locations = letter_locations[0]

    print(global_letter_locations)

    # We then loop through all the lines and only keep those that are in common
    for letter_location in letter_locations[1:]:
        global_letter_locations = set(global_letter_locations).intersection(
            letter_location
        )

    global_letter_locations = sorted(list(set(global_letter_locations)))
    global_letter_locations.append(max_length + 1)

    return global_letter_locations


def split(line: str) -> List[str]:
    """Splits a line of text into a list of characters.

    Args:
        line (str): Line of text to split.

    Returns:
        List[str]: List of characters that make up the line of text.
    """

    # Split the line into a list of characters
    output = list(line)
    return output


def process_line(line: str, widths: List[int]) -> List[str]:
    """Breaks a line apart into cell data according to cell widths.

    Args:
        line (str): Line of text to process.
        widths (List[int]): Widths of each column in the file.

    Returns:
        List: List of strings representing each cell's data.
    """

    char_idx = 0
    output = []

    # Loop through each cell's widths
    for width in widths:
        # grab all the characters from the cell
        data_buffer = line[char_idx : char_idx + width]
        char_idx += width
        # Remove trailing whitespace
        output.append(data_buffer.strip())

    return output


def loads(
    contents: str, padding: str = " ", header: bool = True, output_json: bool = False
) -> Union[List[List], List[Dict]]:
    """Takes a string of a fixed-width file and breaks it apart into the data contained.

    Args:
        contents (str): String fixed-width contents.
        padding (str, optional): Which character takes up the space to create the fixed
            width. Defaults to " ".
        header (bool, optional): Does the file contain a header. Defaults to True.
        output_json (bool, optional): Should a list of dictionaries be returned instead
            of a list of lists. Defaults to False. Requires that 'header' be set to
            True.

    Raises:
        Exception: 'output_json' is True but 'header' is False.

    Returns:
        List[List] | List[Dict]: Either a list of lists or a list of dictionaries that
            represent the extracted data
    """

    lines = contents.split("\n")

    # Remove empty lines
    lines = [line for line in lines if len(line) != 0]

    # Normalize lengths of lines
    lengths = [len(line) for line in lines]
    max_lengths = max(lengths)
    for idx in range(0, len(lines)):  # pylint: disable=C0200
        if lengths[idx] < max_lengths:
            delta = max_lengths - lengths[idx]
            lines[idx] += padding * delta

    # Get the widths of each cell
    # Make sure we use the data to find the widths
    # Because sometimes the headers have spaces in them
    # And that will cause parsing to be weird
    word_locations = get_letter_locations(contents, padding=padding)
    widths = [
        word_locations[i + 1] - word_locations[i]
        for i in range(0, len(word_locations) - 1)
    ]

    if header:
        # Grab the headers if applicable
        headers = process_line(lines[0], widths)
        lines = lines[1:]
    output = []

    # Check that we have a header when requesting JSON output
    # We need the header for the dictionary keys
    if output_json:
        if not header:
            raise Exception(
                "'output_json' requires a header and for 'header' to be set to True"
            )

    for line in lines:
        # Grab the cell data from each line
        data = process_line(line, widths)
        if output_json:
            # Convert to dictionary if applicable
            datum = {}
            for idx, header_key in enumerate(headers):
                datum[header_key] = data[idx]
            # Add dictionary to output
            output.append(datum)
            continue
        # Add list to output
        output.append(data)

    return output
