update-mocks:
	mockery --name=AliasRepository --dir service --output service/mocks --case snake
	mockery --name=MatchRepository --dir service --output service/mocks --case snake
	mockery --name=FootballAPIFixtureRepository --dir service --output service/mocks --case snake
	mockery --name=FootballAPIClient --dir service --output service/mocks --case snake
	mockery --name=TaskScheduler --dir service --output service/mocks --case snake
	mockery --name=Logger --dir service --output service/mocks --case snake

.PHONY: update-mocks