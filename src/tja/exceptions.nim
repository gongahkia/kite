type
  ScrapingError* = object of CatchableError
    url*: string
    statusCode*: int

  NetworkError* = object of ScrapingError
  RateLimitError* = object of ScrapingError
    retryAfter*: int
  ParsingError* = object of ScrapingError
  AuthenticationError* = object of ScrapingError
  DataNotFoundError* = object of ScrapingError
