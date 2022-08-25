local blankChars = ' \r\n\t';

{
  trim(s):: std.stripChars(s, blankChars),
  trimLeft(s):: std.lstripChars(s, blankChars),
  trimRight(s):: std.rstripChars(s, blankChars),
}
