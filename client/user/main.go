package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	pb "example.com/ResolutionSystem/server/GORPC/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FeatureBox struct {
	features      map[string]*pb.Feature
	featuresIdx   map[string]*pb.Feature
	conn          *grpc.ClientConn
	featureClient pb.UserFeatureSeviceClient
	issueClient   pb.UserIssueServiceClient
	common        pb.IssueIntermideateServiceClient
}

func (f *FeatureBox) append(newFeature *pb.Feature) {
	f.features[newFeature.Name.Name] = newFeature
	f.featuresIdx[newFeature.Name.Name] = newFeature
}

func (f *FeatureBox) preloadFeature() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	featureStream, err := f.featureClient.GetAllFeatures(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("error calling function GetAllFeatures: %v", err)
	}

	loaded := make(chan bool)
	go func() {
		for {
			resp, err := featureStream.Recv()
			if err == io.EOF {
				loaded <- true
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			f.append(resp)
		}
	}()
	<-loaded

}

func NewFeatureBox(conn *grpc.ClientConn) *FeatureBox {
	c := pb.NewUserFeatureSeviceClient(conn)
	f := pb.NewUserIssueServiceClient(conn)
	i := pb.NewIssueIntermideateServiceClient(conn)
	features := make(map[string]*pb.Feature)
	idxMap := make(map[string]*pb.Feature)
	ret := &FeatureBox{
		conn:          conn,
		featureClient: c,
		issueClient:   f,
		features:      features,
		featuresIdx:   idxMap,
		common:        i,
	}
	ret.preloadFeature()
	return ret
}

func (f *FeatureBox) SearchByName(name string) (*pb.Feature, error) {
	if resp, ok := f.features[name]; ok {
		return resp, nil
	}
	ctx, close := context.WithTimeout(context.Background(), time.Second)
	defer close()
	resp, err := f.featureClient.GetFeatureByName(ctx, &pb.Name{Name: name})
	if err != nil {
		return nil, err
	}
	f.append(resp)
	return resp, nil
}

func (f *FeatureBox) SearchById(id string) (*pb.Feature, error) {
	if resp, ok := f.featuresIdx[id]; ok {
		return resp, nil
	}
	ctx, close := context.WithTimeout(context.Background(), time.Second)
	defer close()
	resp, err := f.featureClient.GetFeatureById(ctx, &pb.Id{Id: id})
	if err != nil {
		return nil, err
	}
	f.append(resp)
	return resp, nil
}

func (f *FeatureBox) PostIssueId(title string, desc string, id string) (*pb.PostIssueResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	feature, err := f.SearchById(id)

	if err != nil {
		return nil, fmt.Errorf("Function Found with id::#%s Not Found.", id)
	}

	r, r_err := f.issueClient.PostIssue(ctx, &pb.Issue{
		Related:     feature,
		Description: &pb.Description{Description: desc},
		Title:       &pb.Title{Title: title},
	})

	if r_err != nil {
		return nil, r_err
	}
	return r, nil
}

func (f *FeatureBox) PostIssueName(title string, desc string, name string) (*pb.PostIssueResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	feature, err := f.SearchByName(name)

	if err != nil {
		return nil, fmt.Errorf("Function Name %s Not Found.", name)
	}

	r, r_err := f.issueClient.PostIssue(ctx, &pb.Issue{
		Related:     feature,
		Description: &pb.Description{Description: desc},
		Title:       &pb.Title{Title: title},
	})

	if r_err != nil {
		return nil, r_err
	}
	return r, nil
}

func (f *FeatureBox) ForcePostIssue(title string, desc string, related *pb.Feature) (*pb.Issue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	response, err := f.common.CreateIssue(
		ctx,
		&pb.Issue{
			Related:     related,
			Title:       &pb.Title{Title: title},
			Description: &pb.Description{Description: desc},
		},
	)
	if err != nil {
		return nil, err
	}
	return response.GetIssue(), nil
}

func (f *FeatureBox) FindFeatureDes(desc string) ([]*pb.Feature, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	featureStream, err := f.featureClient.GetFeatureByDsc(ctx, &pb.Description{Description: desc})
	if err != nil {
		return nil, err
	}

	result := make([]*pb.Feature, 0)
	loaded := make(chan bool)
	go func() {
		for {
			resp, err := featureStream.Recv()
			if err == io.EOF {
				loaded <- true
				return
			}
			if err != nil {
				loaded <- false
				return
			}
			result = append(result, resp)
		}
	}()
	if <-loaded {
		return result, nil
	}

	return nil, fmt.Errorf("Some Unkown Error Occred")
}

func IssueString(issue *pb.Issue) {
	fmt.Printf("Issue::( ID : #%s\t Title : %s\n\tRelated :: Feature::(\n\t (Id:#%s)\tName:[%s]\n\t Desc:\n\t {%s}\n\t)\n Description:\n\t {%s}\n Solution : {%s}\n)\n",
		issue.Id.Id,
		issue.Title.Title,
		issue.Related.Id.Id,
		issue.Related.Name.Name,
		fmt.Sprint(issue.Related.Description.Description),
		fmt.Sprint(issue.Description.Description),
		issue.Solution,
	)

}

var (
	PORT = os.Getenv("PY_PORT")
	HOST = os.Getenv("HOST")
)

func CLI() {
	target := fmt.Sprintf("%s:%s", HOST, PORT)
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to gRPC server at %s: %v", target, err)
	}
	defer conn.Close()

	fb := NewFeatureBox(conn)

	searchPattern := regexp.MustCompile(`^Search[\s]*:`)
	postIssuePattern := regexp.MustCompile(`^Post[\s]*:`)
	findPattern := regexp.MustCompile(`^Find[\s]*:`)

	instruction := `
+-----------------------------------------------------------------------------+
|                                                                             |
|>>> Search:{Function Name}                                                   |
|                                                                             |
|   ? To Search Function Details by name                                      |
|                                                                             |
|>> Find:{Description}                                                        |
|                                                                             |
|   ? Search / Find Function By description                                   |
|                                                                             |
|>>> Post:{Issue Title}                                                       |
| Function Name >>> {Function Name}                                           |
| Issue Description >>> {What is The Issue you are Facing}                    |
| ...{Result Found}                                                           |
|                                                                             |
| Did You Find You Solution(Y/n) : {You Thinking}                             |
|   ? To Post Issue You are Facing If Simillar Issues are Found Then          |
|   ? Those are Return Else Returns Your Issue is Created with Issue Details  |
|                                                                             |
+-----------------------------------------------------------------------------+
        (ctrl+c or enter) to exit;; (help) to print instruction
`

	reader := bufio.NewReader(os.Stdin)

	fmt.Println(instruction)
	for {
		fmt.Println("=============================================================================")
		fmt.Println("=============================================================================")
		fmt.Print(">>> ")
		input, err := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if err != nil || input == "" {
			break
		}
		switch in := input; {
		case searchPattern.MatchString(in):
			search := searchPattern.Split(in, -1)
			func_name := strings.TrimSpace(search[1])
			var f *pb.Feature
			var err error
			if func_name[0] == '#' {
				f, err = fb.SearchById(func_name)
			} else {
				f, err = fb.SearchByName(func_name)
			}
			if err != nil {
				fmt.Println("Not Found:: Error::", err)
				continue
			}
			fmt.Printf("Found::\n (Id:#%s)\tName:[%s]\n Desc:\n\t{%s}\n", f.Id.Id, f.Name, fmt.Sprint(f.Description.Description))

		case postIssuePattern.MatchString(in):
			issue_title := postIssuePattern.Split(in, -1)
			issue := strings.TrimSpace(issue_title[1])
			fmt.Print("Function Name >>> ")
			func_name, _ := reader.ReadString('\n')
			func_name = strings.TrimSpace(func_name)
			fmt.Print("Issue Description >>> ")
			issue_desc, _ := reader.ReadString('\n')
			issue_desc = strings.TrimSpace(issue_desc)

			var res *pb.PostIssueResponse
			var err error

			if func_name[0] == '#' {
				res, err = fb.PostIssueId(issue, issue_desc, func_name)
			} else {
				res, err = fb.PostIssueName(issue, issue_desc, func_name)
			}

			if err != nil {
				fmt.Printf("Cound Not Post:: Error:: %s\n", err)
				continue
			}

			if res.IsCreated {
				fmt.Printf("Created ::")
				IssueString(res.Created.Issue)
			} else {
				fmt.Printf("Found ::")
				for _, iss := range res.Issues.Issues {
					IssueString(iss)
				}
				fmt.Print("Do You Still want Post : ")
				choise, _ := reader.ReadString('\n')
				choise = strings.TrimSpace(choise)
				if strings.ToLower(choise) == "y" {
					res, err := fb.ForcePostIssue(issue, issue_desc, res.Issues.Issues[0].Related)
					if err != nil {
						fmt.Printf("Couldn't Create Due to : %v\n", err)
						continue
					}
					fmt.Print("Created ::")
					IssueString(res)
				}
			}
		case findPattern.MatchString(in):
			// search := searchPattern.Split(in, -1)
			// desc := strings.TrimSpace(search[1])
			// f, err := fb.FindFeatureDes(desc)

			// if err != nil {

			// }
		case in == "help":
			fmt.Println(instruction)
		default:
			fmt.Printf("Not Found Any Command Name %s\n", input)
			fmt.Println(instruction)
		}
	}
}

func main() {
	CLI()
}
