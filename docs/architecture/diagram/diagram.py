from diagrams import Cluster, Diagram, Edge
from diagrams.onprem.queue import Kafka
from diagrams.onprem.database import PostgreSQL
from diagrams.programming.language import Go
from diagrams.generic.database import SQL

graph_attr = {
    "splines": "line",
    "beautify": "true",
    "fontsize": "16",
    # "bgcolor": "transparent",
}

kafka_cluster_attr = {
    "bgcolor": "lightblue",
    "style": "rounded",
    "pencolor": "grey",
    "fontsize": "16",
    "beautify": "true",
}

database_cluster_attr = {
    "bgcolor": "lightblue",
    "style": "rounded",
    "pencolor": "grey",
    "fontsize": "16",
    "beautify": "true"
}

accruals_cluster_attr = {
    "bgcolor": "lightblue",
    "style": "rounded",
    "pencolor": "grey",
    "fontsize": "16",
    "beautify": "true"
}

gophermart_cluster_attr = {
    "bgcolor": "#EBF3E7",
    "pencolor": "grey",
    "style": "rounded",
    "fontsize": "16",
    "beautify": "true"
}


def main():
    with Diagram("Gophermart Architecture", show=False, direction="TB", graph_attr=graph_attr):
        
        with Cluster("Database", graph_attr=database_cluster_attr):
            pg = PostgreSQL()
            orders = SQL("Orders")
            users = SQL("Users")
            withdraws = SQL("Transactions")
        
        with Cluster("Accrual System Adapter", graph_attr=kafka_cluster_attr):
            new_orders = Go("NEN order handler")
            processed_orders = Go("PENDING order handler")
            dlq_sanitizer = Go("DLQ Handler")
            new = Go("NEW Queue")
            processed = Go("PROCESSING Queue")
            dql = Go("DLQ")
        
        with Cluster("Accruals Service", graph_attr=accruals_cluster_attr):
            haccruals = Go("GET /api/orders/{number}")
        
        with Cluster("Gophermart handlers", graph_attr=gophermart_cluster_attr):
            hregister = Go("POST Register")
            hlogin = Go("POST Login")
            hporders = Go("POST Orders")
            hgorders = Go("GET Orders")
            hbalance = Go("GET Balance")
            hwithdraw = Go("POST Withdraw")
            hwithdraws = Go("GET Withdraws")
            
        # Edges for Kafka and fs
        hporders >> Edge(style="dashed") >> new
        dlq_sanitizer >> Edge(style="dashed") >> new
        dlq_sanitizer >> Edge(style="dashed") >> processed
        dlq_sanitizer << Edge(style="dashed") << dql
        
        # Edges for database
        orders << Edge(style="dashed") << [hporders, hgorders]
        users << Edge(style="dashed") << [hregister, hlogin]
        withdraws << Edge(style="dashed") << hwithdraws
        withdraws << Edge(style="dashed") << hwithdraw
        withdraws << Edge(style="dashed") << hbalance

            
        # Edges for Accruals
        new >> Edge(style="dashed") >> new_orders
        processed << Edge(style="dashed") << new_orders
        dql << Edge(style="dashed") << new_orders
        haccruals >> Edge(style="dashed") >> new_orders
        
        processed >> Edge(style="dashed") >> processed_orders
        processed << Edge(style="dashed") << processed_orders
        dql << Edge(style="dashed") << processed_orders
        haccruals >> Edge(style="dashed") >> processed_orders

if __name__ == "__main__":
    main()
