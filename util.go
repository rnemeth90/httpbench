package httpbench

import (
	"regexp"
	"sort"
	"time"
)

func IsValidURL(url string) bool {
	var urlRegex = regexp.MustCompile(`^(https?|ftp):\/\/[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

func CalculateStatistics(responses []HTTPResponse) Statistics {
	stats := Statistics{}
	stats.TotalCalls = len(responses)
	var totalLatency int64 = 0
	twoHundreds := 0
	threeHundreds := 0
	fourHundreds := 0
	fiveHundreds := 0

	sort.Slice(responses, func(i, j int) bool { return responses[i].Latency < responses[j].Latency })

	stats.FastestRequest = responses[0].Latency
	stats.SlowestRequest = responses[len(responses)-1].Latency

	for _, v := range responses {
		totalLatency = totalLatency + int64(v.Latency)

		switch v.Status {
		case 200:
			twoHundreds++
		case 300, 302:
			threeHundreds++
		case 400:
			fourHundreds++
		case 500:
			fiveHundreds++
		}
	}

	avgLatency := totalLatency / int64(len(responses))
	stats.AvgTimePerRequest = time.Duration(avgLatency)
	stats.TotalTime = time.Duration(totalLatency)

	stats.TwoHundredResponses = twoHundreds
	stats.ThreeHundredResponses = threeHundreds
	stats.FourHundredResponses = fourHundreds
	stats.FiveHundredResponses = fiveHundreds

	return stats
}
