import re

# Read the broken file
with open('pkg/common/fileops/mocks.go', 'r') as f:
    content = f.read()

# Remove the malformed patterns
# Pattern 1: Fix the duplicated return statements
content = re.sub(r'\s+return mockrec\.RecordCall\(mr\.mock\.ctrl, mr\.mock,\nmockrec\.RecordCall\(mr\.mock\.ctrl, mr\.mock, "([^"]+)", reflect\.TypeOf\(\(\*([^)]+)\)\(nil\)\.([^)]+)\)(?:, ([^)]+))?\)', r'\treturn mockrec.RecordCall(mr.mock.ctrl, mr.mock, "\1", reflect.TypeOf((*\2)(nil).\3)\4)', content, flags=re.MULTILINE)

# Pattern 2: Handle cases where args are present
content = re.sub(r'\treturn mockrec\.RecordCall\(mr\.mock\.ctrl, mr\.mock, "([^"]+)", reflect\.TypeOf\(\(\*([^)]+)\)\(nil\)\.([^)]+)\), ([^)]+)\)', r'\treturn mockrec.RecordCall(mr.mock.ctrl, mr.mock, "\1", reflect.TypeOf((*\2)(nil).\3), \4)', content)

# Pattern 3: Handle cases where no args are present (use RecordNoArgsCall)
content = re.sub(r'\treturn mockrec\.RecordCall\(mr\.mock\.ctrl, mr\.mock, "([^"]+)", reflect\.TypeOf\(\(\*([^)]+)\)\(nil\)\.([^)]+)\)\)', r'\treturn mockrec.RecordNoArgsCall(mr.mock.ctrl, mr.mock, "\1", reflect.TypeOf((*\2)(nil).\3))', content)

# Write the fixed content
with open('pkg/common/fileops/mocks.go', 'w') as f:
    f.write(content)

print("Fixed fileops mocks successfully")
EOF < /dev/null