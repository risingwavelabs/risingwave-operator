{
  // 12 is the total row span.
  span(width='half'):: {
    span::: [
      if std.type(width) == 'string' then
        if width == 'full' then 12
        else if width == 'half' then 6
        else if width == 'tierce' then 4
        else if width == 'quarter' then 3
        else error 'width in string must be one of "full", "half", "tierce" and "quarter"'
      else if std.type(width) == 'number' then
        if width > 12 then 12
        else if width > 0 then width
        else error 'width must be a positive number'
      else error 'width must be either a string or a number',
    ],
  },
}
